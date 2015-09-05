package sqlite_backend

import (
	"io/ioutil"
	"sync"

	"gitlab.com/shuhao/towncrier/backend"
)

const (
	standardTestConfigPath = "test_config/standard.conf.json"
	changedTestConfigPath  = "test_config/changed.conf.json"
)

type NotificationSubscriberCombo struct {
	notification backend.Notification
	subscriber   backend.Subscriber
}

type TestNotifier struct {
	*sync.Mutex
	log []*NotificationSubscriberCombo
}

func newTestNotifier() *TestNotifier {
	return &TestNotifier{
		Mutex: &sync.Mutex{},
		log:   []*NotificationSubscriberCombo{},
	}
}

func (n *TestNotifier) Name() string {
	return "testnotify"
}

func (n *TestNotifier) ShouldSendImmediately() bool {
	return false
}

func (n *TestNotifier) Send(notification backend.Notification, subscriber backend.Subscriber) error {
	n.Lock()
	n.log = append(n.log, &NotificationSubscriberCombo{
		notification: notification,
		subscriber:   subscriber,
	})
	n.Unlock()

	return nil
}

var originalTestConfigContent []byte = nil

func changeTestConfig() error {
	var err error
	originalTestConfigContent, err = ioutil.ReadFile(standardTestConfigPath)
	if err != nil {
		return err
	}

	changedTestConfigContent, err := ioutil.ReadFile(changedTestConfigPath)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(standardTestConfigPath, changedTestConfigContent, 0644)
}

func restoreTestConfig() error {
	if originalTestConfigContent == nil {
		panic("original test config not found but trying to restore?")
	}

	return ioutil.WriteFile(standardTestConfigPath, originalTestConfigContent, 0644)
}
