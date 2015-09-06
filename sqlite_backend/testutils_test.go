package sqlite_backend

import (
	"io/ioutil"
	"strings"
	"sync"

	"gitlab.com/shuhao/towncrier/backend"
	. "gopkg.in/check.v1"
)

const (
	standardTestConfigPath = "test_config/standard.conf.json"
	changedTestConfigPath  = "test_config/changed.conf.json"
)

func (s *SQLiteNotificationBackendSuite) checkNotificationEquality(c *C, obtained, expected backend.Notification) {
	obtained.CreatedAt = 0
	obtained.UpdatedAt = 0

	c.Assert(obtained, DeepEquals, expected)
}

type NotificationSubscriberCombo struct {
	notifications []backend.Notification
	subscriber    backend.Subscriber
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

func (n *TestNotifier) Send(notifications []backend.Notification, subscriber backend.Subscriber) error {
	n.Lock()
	n.log = append(n.log, &NotificationSubscriberCombo{
		notifications: notifications,
		subscriber:    subscriber,
	})
	n.Unlock()

	return nil
}

var originalTestConfigContent []byte = nil

func memorizeOriginalConfig() error {
	var err error
	originalTestConfigContent, err = ioutil.ReadFile(standardTestConfigPath)
	return err
}

func changeTestConfig() error {
	err := memorizeOriginalConfig()
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

// Because Go doesn't really have generics
type channelsArray []*Channel

func (a channelsArray) Len() int {
	return len(a)
}

func (a channelsArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a channelsArray) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) == -1
}

type subscribersArray []backend.Subscriber

func (a subscribersArray) Len() int {
	return len(a)
}

func (a subscribersArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a subscribersArray) Less(i, j int) bool {
	return strings.Compare(a[i].UniqueName, a[j].UniqueName) == -1
}
