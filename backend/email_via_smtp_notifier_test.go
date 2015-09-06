package backend

import (
	"net/smtp"
	"time"

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
	notification Notification
	jimmy        Subscriber
}

var _ = Suite(&EmailViaSMTPNotifierSuite{})

const (
	testGmailUsername = "testuser@gmail.com"
	testGmailPassword = "password"
	testSelfEmail     = "testnotifier@example.com"
	testStandardTime  = "2015-09-05T00:15:00Z"
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

func (s *EmailViaSMTPNotifierSuite) SetUpSuite(c *C) {
	s.jimmy = Subscriber{
		UniqueName:  "jimmy",
		Name:        "Jimmy the Cat",
		Email:       "jimmy@the.cat",
		PhoneNumber: "123-456-7890",
	}
}

func (s *EmailViaSMTPNotifierSuite) SetUpTest(c *C) {
	s.sendMailLogs = make([]sendMailParameter, 0)

	s.gmailNotifier = NewEmailViaGmailNotifier(testGmailUsername, testGmailPassword, testSelfEmail)
	s.gmailNotifier.sendMailFunc = s.loggingSendMail

	createdAt, err := time.Parse(time.RFC3339, testStandardTime)
	c.Assert(err, IsNil)

	s.notification = Notification{
		Subject:   "subject",
		Content:   "content abc",
		Channel:   "channel",
		Origin:    "origin",
		Tags:      []string{"tag1", "tag2"},
		Priority:  NormalPriority,
		CreatedAt: createdAt.UnixNano(),
		UpdatedAt: createdAt.UnixNano(),
	}
}

func (s *EmailViaSMTPNotifierSuite) TestSendOne(c *C) {
	err := s.gmailNotifier.Send([]Notification{s.notification}, s.jimmy)
	c.Assert(err, IsNil)

	expectedEmailBody := s.gmailNotifier.convertTextToCRLF(`
From: testnotifier@example.com
To: jimmy@the.cat
Subject: [channel][origin] subject

content abc

Created At: 2015-09-05T00:15:00Z
`) + "\r\n"

	c.Assert(s.sendMailLogs, HasLen, 1)
	c.Assert(s.sendMailLogs[0].from, Equals, testSelfEmail)
	c.Assert(s.sendMailLogs[0].to, DeepEquals, []string{s.jimmy.Email})
	c.Assert(s.sendMailLogs[0].addr, Equals, "smtp.gmail.com:587")
	c.Assert(string(s.sendMailLogs[0].body), Equals, expectedEmailBody)
}

func (s *EmailViaSMTPNotifierSuite) TestSendMany(c *C) {
	err := s.gmailNotifier.Send([]Notification{s.notification, s.notification}, s.jimmy)
	c.Assert(err, IsNil)

	expectedEmailBody := s.gmailNotifier.convertTextToCRLF(`
From: testnotifier@example.com
To: jimmy@the.cat
Subject: [channel] Received 2 notifications

# [origin] subject #

content abc

Created At: 2015-09-05T00:15:00Z

# [origin] subject #

content abc

Created At: 2015-09-05T00:15:00Z
`) + "\r\n"

	c.Assert(s.sendMailLogs, HasLen, 1)
	c.Assert(s.sendMailLogs[0].from, Equals, testSelfEmail)
	c.Assert(s.sendMailLogs[0].to, DeepEquals, []string{s.jimmy.Email})
	c.Assert(s.sendMailLogs[0].addr, Equals, "smtp.gmail.com:587")
	c.Assert(string(s.sendMailLogs[0].body), Equals, expectedEmailBody)
}
