package sqlite_backend

import (
	"sort"
	"testing"

	"gitlab.com/shuhao/towncrier/backend"
	"gitlab.com/shuhao/towncrier/testhelpers"

	. "gopkg.in/check.v1"
)

var logrusTestHook = testhelpers.NewLogrusTestHook()

func Test(t *testing.T) {
	realLogger.Hooks.Add(logrusTestHook)

	TestingT(t)
}

type SQLiteNotificationBackendSuite struct {
	backend  *SQLiteNotificationBackend
	notifier *testhelpers.TestNotifier

	jimmy        backend.Subscriber
	timmy        backend.Subscriber
	bob          backend.Subscriber
	notification backend.Notification
	channel1     *Channel
	channel2     *Channel
}

var _ = Suite(&SQLiteNotificationBackendSuite{})

func (s *SQLiteNotificationBackendSuite) SetUpSuite(c *C) {
	s.jimmy = backend.Subscriber{
		UniqueName:  "jimmy",
		Name:        "Jimmy the Cat",
		Email:       "jimmy@the.cat",
		PhoneNumber: "123-456-7890",
	}

	s.timmy = backend.Subscriber{
		UniqueName:  "timmy",
		Name:        "Timmy the Cat",
		Email:       "timmy@the.cat",
		PhoneNumber: "123-456-7890",
	}

	s.bob = backend.Subscriber{
		UniqueName:  "bob",
		Name:        "Bob the Cat",
		Email:       "bob@the.cat",
		PhoneNumber: "098-765-4321",
	}

	s.notification = backend.Notification{
		Subject:  "subject",
		Content:  "content",
		Origin:   "origin",
		Tags:     []string{"tag1", "tag2"},
		Priority: backend.NormalPriority,
	}

	s.channel1 = &Channel{
		Name:         "Channel1",
		Subscribers:  []string{"jimmy"},
		Notifiers:    []string{"testnotify"},
		TimeToNotify: "@immediately",
	}

	s.channel2 = &Channel{
		Name:         "Channel2",
		Subscribers:  []string{"jimmy", "bob"},
		Notifiers:    []string{"testnotify"},
		TimeToNotify: "@daily",
	}
}

func (s *SQLiteNotificationBackendSuite) SetUpTest(c *C) {
	logrusTestHook.ClearLogs()

	s.backend = backend.GetBackend(BackendName).(*SQLiteNotificationBackend)
	c.Assert(s.backend.Name(), Equals, BackendName)

	err := s.backend.Initialize(":memory:,test_config/standard.conf.json")
	c.Assert(err, IsNil)

	testhelpers.ResetTestDatabase(s.backend.DbMap)

	s.notifier = testhelpers.NewTestNotifier()
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
	c.Assert(s.backend.config.Channels["Channel1"], DeepEquals, s.channel1)
	c.Assert(s.backend.config.Channels["Channel2"], DeepEquals, s.channel2)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationSendImmediately(c *C) {
	notification := s.notification
	notification.Channel = "Channel1"

	err := s.backend.QueueNotification(notification)
	c.Assert(err, IsNil)

	c.Assert(s.notifier.Logs, HasLen, 1)
	c.Assert(s.notifier.Logs[0].Notifications, HasLen, 1)
	s.checkNotificationEquality(c, s.notifier.Logs[0].Notifications[0], notification)
	c.Assert(s.notifier.Logs[0].Subscriber, DeepEquals, s.jimmy)

	notifications := []*Notification{}
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", notification.Channel)
	c.Assert(err, IsNil)
	c.Assert(notifications, HasLen, 1)

	s.checkNotificationEquality(c, notifications[0].Notification, notification)
	c.Assert(notifications[0].Delivered, Equals, true)
}

func (s *SQLiteNotificationBackendSuite) TestQueueUrgentNotificationSendImmediately(c *C) {
	notification := s.notification
	notification.Channel = "Channel2"
	notification.Priority = backend.UrgentPriority

	err := s.backend.QueueNotification(notification)
	c.Assert(err, IsNil)

	c.Assert(s.notifier.Logs, HasLen, 2)
	c.Assert(s.notifier.Logs[0].Notifications, HasLen, 1)
	s.checkNotificationEquality(c, s.notifier.Logs[0].Notifications[0], notification)

	c.Assert(s.notifier.Logs[1].Notifications, HasLen, 1)
	s.checkNotificationEquality(c, s.notifier.Logs[1].Notifications[0], notification)

	subscribersMatched := 0
	for _, subscriber := range []backend.Subscriber{s.bob, s.jimmy} {
		for _, log := range s.notifier.Logs {
			if log.Subscriber.Name == subscriber.Name {
				subscribersMatched++
			}
		}
	}

	c.Assert(subscribersMatched, Equals, 2)

	notifications := []*Notification{}
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", notification.Channel)
	c.Assert(err, IsNil)
	c.Assert(notifications, HasLen, 1)

	s.checkNotificationEquality(c, notifications[0].Notification, notification)
	c.Assert(notifications[0].Delivered, Equals, true)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationDoNotSendImmediately(c *C) {
	notification := s.notification
	notification.Channel = "Channel2"

	err := s.backend.QueueNotification(notification)
	c.Assert(err, IsNil)

	c.Assert(s.notifier.Logs, HasLen, 0)

	notifications := []*Notification{}
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", notification.Channel)
	c.Assert(err, IsNil)
	c.Assert(notifications, HasLen, 1)

	s.checkNotificationEquality(c, notifications[0].Notification, notification)
	c.Assert(notifications[0].Delivered, Equals, false)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationNeverSendNotification(c *C) {
	s.backend.NeverSendNotifications = true
	defer func() { s.backend.NeverSendNotifications = false }()

	notification := s.notification
	notification.Channel = "Channel1"

	err := s.backend.QueueNotification(notification)
	c.Assert(err, IsNil)

	c.Assert(s.notifier.Logs, HasLen, 0)

	notifications := []*Notification{}
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", notification.Channel)
	c.Assert(err, IsNil)
	c.Assert(notifications, HasLen, 1)

	s.checkNotificationEquality(c, notifications[0].Notification, notification)
	c.Assert(notifications[0].Delivered, Equals, false)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotificationWithInvalidChannelWillNotSave(c *C) {
	notification := s.notification
	notification.Channel = "invalid-channel"

	err := s.backend.QueueNotification(notification)
	c.Assert(err, NotNil)
	channelNotFoundErr, ok := err.(backend.ChannelNotFound)
	c.Assert(ok, Equals, true)
	c.Assert(channelNotFoundErr.ChannelName, Equals, "invalid-channel")

	notifications := []*Notification{}
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", notification.Channel)
	c.Assert(err, IsNil)
	c.Assert(notifications, HasLen, 0)
}

func (s *SQLiteNotificationBackendSuite) TestName(c *C) {
	c.Assert(s.backend.Name(), Equals, BackendName)
}

func (s *SQLiteNotificationBackendSuite) TestGetChannels(c *C) {
	channels := channelsArray(s.backend.GetChannels())
	c.Assert(channels, HasLen, 2)

	sort.Sort(channels)

	c.Assert(channels[0], DeepEquals, s.channel1)
	c.Assert(channels[1], DeepEquals, s.channel2)
}

func (s *SQLiteNotificationBackendSuite) TestGetSubscribers(c *C) {
	subscribers := subscribersArray(s.backend.GetSubscribers())
	c.Assert(subscribers, HasLen, 2)

	sort.Sort(subscribers)

	c.Assert(subscribers[0], DeepEquals, s.bob)
	c.Assert(subscribers[1], DeepEquals, s.jimmy)
}
