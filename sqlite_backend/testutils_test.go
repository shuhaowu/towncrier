package sqlite_backend

import (
	"sync"

	"gitlab.com/shuhao/towncrier/backend"
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
