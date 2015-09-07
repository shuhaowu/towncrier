package webreceiver

import (
	"testing"

	"gitlab.com/shuhao/towncrier/backend"
	"gitlab.com/shuhao/towncrier/sqlite_backend"
	"gitlab.com/shuhao/towncrier/testhelpers"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type WebReceiverAppSuite struct {
	app      *App
	config   ReceiverConfig
	backend  *sqlite_backend.SQLiteNotificationBackend
	notifier *testhelpers.TestNotifier
}

var _ = Suite(&WebReceiverAppSuite{})

func (s *WebReceiverAppSuite) SetUpSuite(c *C) {
	s.backend = backend.GetBackend(sqlite_backend.BackendName).(*sqlite_backend.SQLiteNotificationBackend)
	err := s.backend.Initialize(":memory:,../sqlite_backend/test_config/standard.conf.json")
	c.Assert(err, IsNil)

	s.config = ReceiverConfig{
		ListenHost: "127.0.0.1",
		ListenPort: 3921,
		PathPrefix: "/receiver",
		Tokens: map[string]string{
			"abc": "abc_client",
		},
	}
}

func (s *WebReceiverAppSuite) SetUpTest(c *C) {
	s.notifier = testhelpers.NewTestNotifier()
	backend.ClearAllNotifiers()
	backend.RegisterNotifier(s.notifier)

	s.app = NewApp(s.backend, s.config)
}

func (s *WebReceiverAppSuite) TestPostNotificationNotAuthenticated(c *C) {

}

func (s *WebReceiverAppSuite) TestPostNotificationInvalidData(c *C) {

}

func (s *WebReceiverAppSuite) TestPostNotificationChannelNotFound(c *C) {

}

func (s *WebReceiverAppSuite) TestPostNotificationSuccess(c *C) {

}
