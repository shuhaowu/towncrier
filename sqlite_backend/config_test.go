package sqlite_backend

import (
	"bytes"
	"io/ioutil"
	"time"

	. "gopkg.in/check.v1"
)

func (s *SQLiteNotificationBackendSuite) TestConfigReloadWillBlockIfLocked(c *C) {
	s.backend.config.Lock()

	timedout := false

	go func() {
		// :troll:
		time.Sleep(200 * time.Millisecond)
		s.backend.config.Unlock()
		timedout = true
	}()

	err := s.backend.config.Reload()
	c.Assert(timedout, Equals, true)
	c.Assert(err, IsNil)
}

func (s *SQLiteNotificationBackendSuite) TestConfigReloadWillFailIfTimeToNotifyIsWrong(c *C) {
	err := memorizeOriginalConfig()
	c.Assert(err, IsNil)

	modifiedConfig := bytes.Replace(originalTestConfigContent, []byte(ChannelSendImmediately), []byte("meh"), -1)

	err = ioutil.WriteFile(standardTestConfigPath, modifiedConfig, 0644)
	c.Assert(err, IsNil)
	defer restoreTestConfig()

	err = s.backend.config.Reload()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "channel 'Channel1' has an invalid TimeToNotify")
}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeMinutely(c *C) {

	channel := &Channel{
		Name:         "testchannel",
		TimeToNotify: "* * * * * *",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	// Minutely pretty much should always be true.
	currentTime, err := time.Parse(time.RFC3339, "2015-09-05T16:38:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:38:01Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:38:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:38:59Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)
}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeHourly(c *C) {
	channel := &Channel{
		Name:         "testchannel",
		TimeToNotify: "@hourly",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err := time.Parse(time.RFC3339, "2015-09-05T16:38:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:58:59Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:59:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T17:00:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T17:02:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	channel = &Channel{
		Name:         "testchannel",
		TimeToNotify: "15 * * * * *",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:38:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:13:59Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:14:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:14:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:15:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T16:20:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)
}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeDaily(c *C) {
	channel := &Channel{
		Name:         "testchannel",
		TimeToNotify: "@daily",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err := time.Parse(time.RFC3339, "2015-09-05T16:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T23:59:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T23:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-06T00:00:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-06T00:02:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	channel = &Channel{
		Name:         "testchannel",
		TimeToNotify: "15 1 * * * *",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T23:59:59Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T01:14:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T01:14:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T01:15:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T01:20:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)
}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeWeekly(c *C) {
	channel := &Channel{
		Name:         "testchannel",
		TimeToNotify: "@weekly",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err := time.Parse(time.RFC3339, "2015-09-05T16:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	// This is a Saturday -> Sunday transition
	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T23:59:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T23:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-06T00:00:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-06T00:02:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	channel = &Channel{
		Name:         "testchannel",
		TimeToNotify: "15 3 * * 1 *", // 3:15 AM on a Monday
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T03:14:59Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-07T03:14:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-07T03:14:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-07T03:15:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-07T03:17:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)
}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeMonthly(c *C) {
	channel := &Channel{
		Name:         "testchannel",
		TimeToNotify: "@monthly",
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err := time.Parse(time.RFC3339, "2015-09-30T16:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	// This is a Saturday -> Sunday transition
	currentTime, err = time.Parse(time.RFC3339, "2015-09-30T23:59:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-30T23:59:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-10-01T00:00:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-10-01T00:02:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	channel = &Channel{
		Name:         "testchannel",
		TimeToNotify: "15 3 5 * * *", // 3:15 AM on day 5 of a month
	}

	c.Assert(channel.IsTimeToNotifyValid(), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-02T03:14:59Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T03:14:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T03:14:30Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T03:15:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, true)

	currentTime, err = time.Parse(time.RFC3339, "2015-09-05T03:17:00Z")
	c.Assert(err, IsNil)
	c.Assert(channel.ShouldSendNowGivenTime(currentTime), Equals, false)
}
