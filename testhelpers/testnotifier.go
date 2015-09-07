package testhelpers

import (
	"sync"

	"gitlab.com/shuhao/towncrier/backend"
)

type NotificationSubscriberCombo struct {
	Notifications []backend.Notification
	Subscriber    backend.Subscriber
}

type TestNotifier struct {
	*sync.Mutex
	Logs []*NotificationSubscriberCombo
}

func NewTestNotifier() *TestNotifier {
	return &TestNotifier{
		Mutex: &sync.Mutex{},
		Logs:  []*NotificationSubscriberCombo{},
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
	n.Logs = append(n.Logs, &NotificationSubscriberCombo{
		Notifications: notifications,
		Subscriber:    subscriber,
	})
	n.Unlock()

	return nil
}
