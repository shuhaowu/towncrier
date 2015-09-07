package webreceiver

import (
	"testing"

	"gitlab.com/shuhao/towncrier/sqlite_backend"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type WebReceiverAppSuite struct {
	app     *App
	config  ReceiverConfig
	backend *sqlite_backend.SQLiteNotificationBackend
}

var _ = Suite(&WebReceiverAppSuite{})

func (s *WebReceiverAppSuite) SetUpTest(c *C) {

}
