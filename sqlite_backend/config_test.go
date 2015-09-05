package sqlite_backend

import (
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

}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeMinutely(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeHourly(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeDaily(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeWeekly(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeMonthly(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestChannelShouldSendGivenTimeYearly(c *C) {

}
