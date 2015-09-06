package backend

import (
	"net/smtp"

	. "gopkg.in/check.v1"
)

type sendMailParameter struct {
	addr string
	a    smtp.Auth
	from string
	to   []string
	body []byte
}

type EmailViaSMTPNotifierSuite struct {
	gmailNotifier *EmailViaSMTPNotifier

	sendMailLogs []sendMailParameter
}

var _ = Suite(&EmailViaSMTPNotifierSuite{})

const (
	testGmailUsername = "testuser@gmail.com"
	testGmailPassword = "password"
	testSelfEmail     = "testnotifier@example.com"
)

func (s *EmailViaSMTPNotifierSuite) loggingSendMail(addr string, a smtp.Auth, from string, to []string, body []byte) error {
	s.sendMailLogs = append(s.sendMailLogs, sendMailParameter{
		addr: addr,
		a:    a,
		from: from,
		to:   to,
		body: body,
	})

	return nil
}

func (s *EmailViaSMTPNotifierSuite) SetUpTest(c *C) {
	s.sendMailLogs = make([]sendMailParameter, 0)

	s.gmailNotifier = NewEmailViaGmailNotifier(testGmailUsername, testGmailPassword, testSelfEmail)
	s.gmailNotifier.sendMailFunc = s.loggingSendMail
}

func (s *EmailViaSMTPNotifierSuite) TestSendOne(c *C) {

}

func (s *EmailViaSMTPNotifierSuite) TestSendMany(c *C) {

}
