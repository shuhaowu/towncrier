package backend

import (
	"bytes"
	"fmt"
	"net/smtp"
	"regexp"
	"strings"
	"text/template"
)

const emailTemplateText = "From: {{.From}}\r\n" +
	"To: {{.To}}\r\n" +
	"Subject: {{.Subject}}\r\n" +
	"\r\n" +
	"{{.Body}}\r\n"

var emailTemplate = template.Must(template.New("email").Parse(emailTemplateText))

var newlineRegex = regexp.MustCompile("\r?\n")

type emailTemplateData struct {
	From    string
	To      string
	Subject string
	Body    string
}

type EmailViaSMTPNotifier struct {
	Hostname      string
	Port          int
	SelfEmailAddr string
	Authenticator smtp.Auth
}

func NewEmailViaSMTPNotifier(hostname string, port int, authenticator smtp.Auth, selfEmail string) *EmailViaSMTPNotifier {
	return &EmailViaSMTPNotifier{
		Hostname:      hostname,
		Port:          port,
		Authenticator: authenticator,
		SelfEmailAddr: selfEmail,
	}
}

func NewEmailViaGmailNotifier(username string, password string, selfEmail string) *EmailViaSMTPNotifier {
	return &EmailViaSMTPNotifier{
		Hostname:      "smtp.gmail.com",
		Port:          587,
		Authenticator: smtp.PlainAuth("", username, password, "smtp.gmail.com"),
		SelfEmailAddr: selfEmail,
	}
}

func (n *EmailViaSMTPNotifier) Name() string {
	return "EmailViaSMTP"
}

func (n *EmailViaSMTPNotifier) Send(notifications []Notification, subscriber Subscriber) error {
	if len(notifications) == 1 {
		return n.SendOne(notifications[0], subscriber)
	} else {
		return n.SendMany(notifications, subscriber)
	}
}

func (n *EmailViaSMTPNotifier) convertTextToCRLF(text string) string {
	text = strings.TrimSpace(text)
	lines := newlineRegex.Split(text, -1)
	return strings.Join(lines, "\r\n")
}

func (n *EmailViaSMTPNotifier) SendOne(notification Notification, subscriber Subscriber) error {
	data := emailTemplateData{
		From:    n.SelfEmailAddr,
		To:      subscriber.Email,
		Subject: fmt.Sprintf("[%s][%s] %s", notification.Channel, notification.Origin, notification.Subject),
		Body:    notification.Content,
	}

	emailBuf := &bytes.Buffer{}
	err := emailTemplate.Execute(emailBuf, data)
	if err != nil {
		return err
	}

	return smtp.SendMail(fmt.Sprintf("%s:%d", n.Hostname, n.Port), n.Authenticator, n.SelfEmailAddr, []string{subscriber.Email}, emailBuf.Bytes())
}

func (n *EmailViaSMTPNotifier) SendMany(notifications []Notification, subscriber Subscriber) error {
	return nil
}
