package sqlite_backend

import (
	"io/ioutil"
	"sort"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	. "gopkg.in/check.v1"
)

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
}
