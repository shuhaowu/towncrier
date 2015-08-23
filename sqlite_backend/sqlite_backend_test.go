package sqlite_backend

import (
	"testing"

	"gitlab.com/shuhao/towncrier/backend"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type SQLiteNotificationBackendSuite struct {
	backend *SQLiteNotificationBackend
}

var _ = Suite(&SQLiteNotificationBackendSuite{})

func (s *SQLiteNotificationBackendSuite) SetUpTest(c *C) {
	s.backend = &SQLiteNotificationBackend{
		quitChannel: make(chan struct{}),
	}
}

func (s *SQLiteNotificationBackendSuite) TestBackendRegistered(c *C) {
	sqliteBackend := backend.GetBackend(BackendName)
	c.Assert(sqliteBackend.Name(), Equals, BackendName)
}

func (s *SQLiteNotificationBackendSuite) TestBackendInitialize(c *C) {
	err := s.backend.Initialize(":memory:,test_config/standard.conf.json")
	c.Assert(err, IsNil)

	c.Assert(s.backend.config, NotNil)
	c.Assert(s.backend.config.Subscribers, HasLen, 2)

	jimmy := &backend.Subscriber{
		UniqueName:  "jimmy",
		Name:        "Jimmy the Cat",
		Email:       "jimmy@the.cat",
		PhoneNumber: "123-456-7890",
	}

	bob := &backend.Subscriber{
		UniqueName:  "bob",
		Name:        "Bob the Cat",
		Email:       "bob@the.cat",
		PhoneNumber: "098-765-4321",
	}
	c.Assert(s.backend.config.Subscribers["jimmy"], DeepEquals, jimmy)
	c.Assert(s.backend.config.Subscribers["bob"], DeepEquals, bob)
}

func (s *SQLiteNotificationBackendSuite) TestQueueNotification(c *C) {

}

func (s *SQLiteNotificationBackendSuite) TestName(c *C) {
	c.Assert(s.backend.Name(), Equals, BackendName)
}
