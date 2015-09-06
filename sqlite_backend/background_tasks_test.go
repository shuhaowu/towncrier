package sqlite_backend

import (
	"io/ioutil"
	"sort"
	"sync"
	"time"

	"gitlab.com/shuhao/towncrier/backend"
	"gitlab.com/shuhao/towncrier/testhelpers"

	"github.com/Sirupsen/logrus"

	. "gopkg.in/check.v1"
)

const testBackgroundTaskTimeout = 5 * time.Second

func (s *SQLiteNotificationBackendSuite) TestStartsConfigReloaderAndNotificationDelivery(c *C) {
	wg := &sync.WaitGroup{}
	s.backend.Start(wg)
	s.backend.BlockUntilReady()
	defer s.backend.Shutdown()

	c.Assert(logrusTestHook.Logs[logrus.InfoLevel], HasLen, 2)
	entries := make(map[string]bool)

	for _, entry := range logrusTestHook.Logs[logrus.InfoLevel] {
		entries[entry.Message] = true
	}

	c.Assert(entries["started config reloader"], Equals, true)
	c.Assert(entries["started notification delivery"], Equals, true)

	err := changeTestConfig()
	c.Assert(err, IsNil)
	defer restoreTestConfig()

	s.backend.ForceConfigReload()

	// I'M SORRY
	time.Sleep(200 * time.Millisecond)

	channels := channelsArray(s.backend.GetChannels())
	subscribers := subscribersArray(s.backend.GetSubscribers())

	sort.Sort(channels)
	sort.Sort(subscribers)

	c.Assert(channels, HasLen, 2)
	c.Assert(channels[0].Name, Equals, "Channel1")
	c.Assert(channels[0].Subscribers, DeepEquals, []string{"timmy"})

	c.Assert(subscribers, HasLen, 2)
	c.Assert(subscribers[0], DeepEquals, s.bob)
	c.Assert(subscribers[1], DeepEquals, s.timmy)
}

func (s *SQLiteNotificationBackendSuite) TestDoConfigReloadLogOnError(c *C) {
	err := memorizeOriginalConfig()
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(standardTestConfigPath, []byte("faulty data"), 0644)
	c.Assert(err, IsNil)
	defer restoreTestConfig()

	s.backend.doConfigReloadLogIfError()

	c.Assert(logrusTestHook.Logs[logrus.ErrorLevel], HasLen, 1)

	channels := channelsArray(s.backend.GetChannels())
	subscribers := subscribersArray(s.backend.GetSubscribers())

	sort.Sort(channels)
	sort.Sort(subscribers)

	c.Assert(channels, HasLen, 2)
	c.Assert(channels[0], DeepEquals, s.channel1)
	c.Assert(channels[1], DeepEquals, s.channel2)

	c.Assert(subscribers, HasLen, 2)
	c.Assert(subscribers[0], DeepEquals, s.bob)
	c.Assert(subscribers[1], DeepEquals, s.jimmy)
}

func (s *SQLiteNotificationBackendSuite) TestDeliverNotificationsWillNotWithoutNotifications(c *C) {
	currentTime, err := time.Parse(time.RFC3339, "2015-09-05T23:59:59Z")
	c.Assert(err, IsNil)
	s.backend.deliverNotificationsLogIfError(currentTime)

	c.Assert(s.notifier.log, HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.InfoLevel], HasLen, 1)
	c.Assert(logrusTestHook.Logs[logrus.WarnLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.ErrorLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.InfoLevel][0].Message, Equals, "attempting to deliver notifications")

	notification := &Notification{
		Notification: s.notification,
		Delivered:    true,
	}
	notification.Channel = "Channel2"

	err = notification.insert(s.backend.DbMap)
	c.Assert(err, IsNil)

	logrusTestHook.ClearLogs()
	s.backend.deliverNotificationsLogIfError(currentTime)
	// Since we use asynchronous sending. In case it sent we want to catch it.
	time.Sleep(200 * time.Millisecond)

	c.Assert(s.notifier.log, HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.InfoLevel], HasLen, 1)
	c.Assert(logrusTestHook.Logs[logrus.WarnLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.ErrorLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.InfoLevel][0].Message, Equals, "attempting to deliver notifications")

	notification.Delivered = false
	_, err = s.backend.Update(notification)
	c.Assert(err, IsNil)

	logrusTestHook.ClearLogs()
	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T23:55:59Z")
	c.Assert(err, IsNil)

	s.backend.deliverNotificationsLogIfError(currentTime)
	// Since we use asynchronous sending. In case it sent we want to catch it.
	time.Sleep(200 * time.Millisecond)

	c.Assert(s.notifier.log, HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.InfoLevel], HasLen, 1)
	c.Assert(logrusTestHook.Logs[logrus.WarnLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.ErrorLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.InfoLevel][0].Message, Equals, "attempting to deliver notifications")
}

func (s *SQLiteNotificationBackendSuite) TestDeliverNotificationsDelivers(c *C) {
	notification := s.notification
	notification.Channel = "Channel2"

	localNotification := &Notification{
		Notification: notification,
		Delivered:    false,
	}

	localNotification2 := &Notification{
		Notification: notification,
		Delivered:    true,
	}

	err := s.backend.DbMap.Insert(localNotification, localNotification2)
	c.Assert(err, IsNil)

	currentTime, err := time.Parse(time.RFC3339, "2015-09-05T23:59:59Z")
	c.Assert(err, IsNil)
	s.backend.deliverNotificationsLogIfError(currentTime)

	timedout := testhelpers.BlockUntilSatisfiedOrTimeout(func() bool {
		return len(s.notifier.log) >= 2
	}, testBackgroundTaskTimeout)

	c.Assert(timedout, Equals, false)

	c.Assert(s.notifier.log, HasLen, 2)
	c.Assert(s.notifier.log[0].notifications, HasLen, 1)
	s.checkNotificationEquality(c, s.notifier.log[0].notifications[0], notification)

	c.Assert(s.notifier.log[1].notifications, HasLen, 1)
	s.checkNotificationEquality(c, s.notifier.log[1].notifications[0], notification)

	subscribersMatched := 0
	for _, subscriber := range []backend.Subscriber{s.bob, s.jimmy} {
		for _, log := range s.notifier.log {
			if log.subscriber.Name == subscriber.Name {
				subscribersMatched++
			}
		}
	}

	c.Assert(subscribersMatched, Equals, 2)

	c.Assert(logrusTestHook.Logs[logrus.WarnLevel], HasLen, 0)
	c.Assert(logrusTestHook.Logs[logrus.ErrorLevel], HasLen, 0)
}
