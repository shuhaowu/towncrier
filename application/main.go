package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"gitlab.com/shuhao/towncrier/backend"
	"gitlab.com/shuhao/towncrier/webreceiver"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/grace/gracehttp"
)

// Flag variables
var configPath string

// Configuration variables
var applicationConfig *ApplicationConfig

var realLogger = logrus.New()
var logger = realLogger.WithField("component", "main")

func init() {
	flag.StringVar(&configPath, "config", "", "the master config file for the server")
	flag.Parse()

	if configPath == "" || !pathExists(configPath) {
		logger.WithField("path", configPath).Panic("config path not valid")
	}

	var err error
	applicationConfig, err = NewApplicationConfig(configPath)
	if err != nil {
		logger.WithField("error", err).Panic("cannot parse application config")
	}
}

func main() {
	wg := &sync.WaitGroup{}

	notificationBackend := backend.GetBackend(applicationConfig.BackendName)
	err := notificationBackend.Initialize(applicationConfig.BackendOpenString)
	if err != nil {
		logger.WithField("error", err).Panic("cannot initialize backend")
	}

	notificationBackend.Start(wg)
	notificationBackend.BlockUntilReady()
	receiverApp = webreceiver.NewApp(notificationBackend, applicationConfig.Receiver)

	receiverServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", applicationConfig.Receiver.ListenHost, applicationConfig.Receiver.ListenPort),
		Handler: receiverApp,
	}

	gracehttp.Serve(receiverServer)
}
