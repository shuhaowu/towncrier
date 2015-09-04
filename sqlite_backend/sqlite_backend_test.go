package sqlite_backend

import (
	"testing"

	"gitlab.com/shuhao/towncrier/backend"
	"gitlab.com/shuhao/towncrier/testhelpers"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type SQLiteNotificationBackendSuite struct {
	backend  *SQLiteNotificationBackend
	notifier *TestNotifier

	jimmy backend.Subscriber
	bob   backend.Subscriber
}

var _ = Suite(&SQLiteNotificationBackendSuite{})

func (s *SQLiteNotificationBackendSuite) SetUpSuite(c *C) {
	s.jimmy = backend.Subscriber{
		UniqueName:  "jimmy",
		Name:        "Jimmy the Cat",
		Email:       "jimmy@the.cat",
		PhoneNumber: "123-456-7890",
	}

	s.bob = backend.Subscriber{
		UniqueName:  "bob",
		Name:        "Bob the Cat",
		Email:       "bob@the.cat",
		PhoneNumber: "098-765-4321",
	}
}

func (s *SQLiteNotificationBackendSuite) SetUpTest(c *C) {
	s.backend = backend.GetBackend(BackendName).(*SQLiteNotificationBackend)
	c.Assert(s.backend.Name(), Equals, BackendName)

	err := s.backend.Initialize(":memory:,test_config/standard.conf.json")
	c.Assert(err, IsNil)

	testhelpers.ResetTestDatabase(s.backend.DbMap)

	s.notifier = newTestNotifier()
	backend.ClearAllNotifiers()
	backend.RegisterNotifier(s.notifier)
}

func (s *SQLiteNotificationBackendSuite) TestBackendInitialize(c *C) {
	s.backend = &SQLiteNotificationBackend{
		NeverSendNotifications: false,
		quitChannel:            make(chan struct{}),
	}

	err := s.backend.Initialize(":memory:,test_config/standard.conf.json")
	c.Assert(err, IsNil)

	c.Assert(s.backend.config, NotNil)
	c.Assert(s.backend.config.Subscribers, HasLen, 2)
	c.Assert(s.backend.config.Subscribers["jimmy"], DeepEquals, s.jimmy)
	c.Assert(s.backend.config.Subscribers["bob"], DeepEquals, s.bob)

	c.Assert(s.backend.config.Channels, HasLen, 2)

	channel1 := &Channel{
		Name:         "Channel1",
		Subscribers:  []string{"jimmy"},
		Notifiers:    []string{"testnotify"},
		TimeToNotify: "@immediately",
	}

	channel2 := &Channel{
		Name:         "Channel2",
		Subscribers:  []string{"jimmy", "bob"},
		Notifiers:    []string{"testnotify"},
		TimeToNotify: "@daily",
	}
	c.Assert(s.backend.config.Channels["Channel1"], DeepEquals, channel1)
	c.Assert(s.backend.config.Channels["Channel2"], DeepEquals, channel2)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationSendImmediately(c *C) {
	notification := backend.Notification{
		Subject:  "subject",
		Content:  "content",
		Channel:  "Channel1",
		Origin:   "origin",
		Tags:     []string{"tag1", "tag2"},
		Priority: backend.NormalPriority,
	}

	err := s.backend.QueueNotification(notification)
	c.Assert(err, IsNil)

	c.Assert(s.notifier.log, HasLen, 1)
	c.Assert(s.notifier.log[0].notification, DeepEquals, notification)
	c.Assert(s.notifier.log[0].subscriber, DeepEquals, s.jimmy)

	notifications := []*Notification{}
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", "Channel1")
	c.Assert(err, IsNil)
	c.Assert(notifications, HasLen, 1)

	c.Assert(notifications[0].Notification, DeepEquals, notification)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationDoNotSendImmediately(c *C) {
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationNeverSendNotification(c *C) {
}

func (s *SQLiteNotificationBackendSuite) TestName(c *C) {
	c.Assert(s.backend.Name(), Equals, BackendName)
}

func (s *SQLiteNotificationBackendSuite) TestStartShutdown(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestGetChannelsGetSubscribers(c *C) {

}
