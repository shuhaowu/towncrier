package webreceiver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	server   *httptest.Server
}

var _ = Suite(&WebReceiverAppSuite{})

func (s *WebReceiverAppSuite) SetUpSuite(c *C) {

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

	s.backend = backend.GetBackend(sqlite_backend.BackendName).(*sqlite_backend.SQLiteNotificationBackend)
	err := s.backend.Initialize(":memory:,../sqlite_backend/test_config/standard.conf.json")
	c.Assert(err, IsNil)

	// TODO: start the backend

	testhelpers.ResetTestDatabase(s.backend.DbMap)

	s.app = NewApp(s.backend, s.config)
	s.server = httptest.NewServer(s.app)
}

func (s *WebReceiverAppSuite) TearDownTest(c *C) {
	s.server.Close()
}

func (s *WebReceiverAppSuite) url(path string) string {
	return s.server.URL + path
}

func (s *WebReceiverAppSuite) postNotification(channel, token, subject, content, tags, priority string) (*http.Response, error) {
	client := &http.Client{}
	contentBuf := bytes.NewBufferString(content)
	req, err := http.NewRequest("POST", s.url("/receiver/notifications/"+channel), contentBuf)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", "Token token="+token)
	}
	req.Header.Add("X-Towncrier-Subject", subject)
	req.Header.Add("X-Towncrier-Tags", tags)
	req.Header.Add("X-Towncrier-Priority", priority)

	return client.Do(req)
}

func (s *WebReceiverAppSuite) TestPostNotificationNotAuthenticated(c *C) {
	resp, err := s.postNotification("Channel1", "invalid-token", "subject", "content", "tag1,tag2", "normal")
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusForbidden)

	resp, err = s.postNotification("Channel1", "", "subject", "content", "tag1,tag2", "normal")
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusForbidden)
}

func (s *WebReceiverAppSuite) TestPostNotificationChannelNotFound(c *C) {
	resp, err := s.postNotification("InvalidChannel", "abc", "subject", "content", "tag1,tag2", "normal")
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusNotFound)
}

func (s *WebReceiverAppSuite) TestPostNotificationSuccess(c *C) {
	resp, err := s.postNotification("Channel1", "abc", "subject", "content", "tag1,tag2", "normal")
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusAccepted)

	timedout := testhelpers.BlockUntilSatisfiedOrTimeout(func() bool {
		return len(s.notifier.Logs) >= 1
	}, 5*time.Second)
	c.Assert(timedout, Equals, false)

	var notifications []*sqlite_backend.Notification
	_, err = s.backend.Select(&notifications, "SELECT * FROM notifications WHERE Channel = ?", "Channel1")
	c.Assert(err, IsNil)

	c.Assert(notifications, HasLen, 1)
	c.Assert(notifications[0].Subject, Equals, "subject")
	c.Assert(notifications[0].Content, Equals, "content")
	c.Assert(notifications[0].Tags, DeepEquals, []string{"tag1", "tag2"})
	c.Assert(notifications[0].Priority, Equals, backend.NormalPriority)
	c.Assert(notifications[0].Channel, Equals, "Channel1")
	c.Assert(notifications[0].Origin, Equals, "abc_client")
	c.Assert(notifications[0].Delivered, Equals, true)

	c.Assert(s.notifier.Logs, HasLen, 1)
	c.Assert(s.notifier.Logs[0].Notifications, HasLen, 1)
	c.Assert(s.notifier.Logs[0].Notifications[0].Subject, Equals, "subject")
	c.Assert(s.notifier.Logs[0].Subscriber.UniqueName, Equals, "jimmy")
}
