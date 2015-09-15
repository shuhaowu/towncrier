package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"

	"gitlab.com/shuhao/towncrier/backend"
	_ "gitlab.com/shuhao/towncrier/sqlite_backend"
	"gitlab.com/shuhao/towncrier/webreceiver"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/grace/gracehttp"
)

// Flag variables
var configPath string
var pidPath string

// Configuration variables
var applicationConfig *ApplicationConfig

var realLogger = logrus.New()
var logger = realLogger.WithField("component", "main")

func init() {
	flag.StringVar(&configPath, "config", "", "the master config file for the server")
	flag.StringVar(&pidPath, "pidfile", "", "the pid file path for the server")
	flag.Parse()

	if configPath == "" || !pathExists(configPath) {
		logger.WithField("path", configPath).Panic("config path not valid")
	}

	var err error
	pid := os.Getpid()
	if pidPath != "" {
		logger.WithFields(logrus.Fields{
			"path": pidPath,
			"pid":  pid,
		}).Info("creating pid file")

		err = ioutil.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644)
		if err != nil {
			logger.WithField("error", err).Panic("cannot write to pid file")
		}
	}

	applicationConfig, err = NewApplicationConfig(configPath)
	if err != nil {
		logger.WithField("error", err).Panic("cannot parse application config")
	}

	applicationConfig.Notifiers.HookAllNotifiers()
}

func main() {
	wg := &sync.WaitGroup{}

	notificationBackend := backend.GetBackend(applicationConfig.BackendName)
	err := notificationBackend.Initialize(applicationConfig.BackendOpenString)
	if err != nil {
		logger.WithField("error", err).Panic("cannot initialize backend")
	}

	notificationBackend.Start(wg)
	defer func() {
		notificationBackend.Shutdown()
		wg.Wait()
	}()

	notificationBackend.BlockUntilReady()
	receiverApp := webreceiver.NewApp(notificationBackend, applicationConfig.Receiver)

	receiverServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", applicationConfig.Receiver.ListenHost, applicationConfig.Receiver.ListenPort),
		Handler: receiverApp,
	}

	gracehttp.Serve(receiverServer)
}
