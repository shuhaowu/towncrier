package backend

import (
	"bytes"
	"fmt"
	"net/smtp"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const emailTemplateText = "From: {{.From}}\r\n" +
	"To: {{.To}}\r\n" +
	"Subject: {{.Subject}}\r\n" +
	"\r\n" +
	"{{.Body}}\r\n"

const singleNotificationTemplateText = `# [{{.Origin}}] {{.Subject}} #

{{.Content}}

Created At: {{.Timestamp}}
`

var emailTemplate = template.Must(template.New("email").Parse(emailTemplateText))
var singleNotificationTemplate = template.Must(template.New("singleNotification").Parse(singleNotificationTemplateText))

var newlineRegex = regexp.MustCompile("\r?\n")

type emailTemplateData struct {
	From    string
	To      string
	Subject string
	Body    string
}

func (d emailTemplateData) emailBytes() ([]byte, error) {
	emailBuf := &bytes.Buffer{}
	err := emailTemplate.Execute(emailBuf, d)
	if err != nil {
		return nil, err
	}

	return emailBuf.Bytes(), nil
}

type notificationData struct {
	Origin    string
	Subject   string
	Content   string
	Timestamp string
}

func newNotificationData(notification Notification) notificationData {
	return notificationData{
		Origin:    notification.Origin,
		Subject:   notification.Subject,
		Content:   notification.Content,
		Timestamp: time.Unix(0, notification.CreatedAt).UTC().Format(time.RFC3339),
	}
}

type EmailViaSMTPNotifier struct {
	Hostname      string
	Port          int
	SelfEmailAddr string
	Authenticator smtp.Auth

	// For testing...
	sendMailFunc func(string, smtp.Auth, string, []string, []byte) error
}

func printMail(addr string, a smtp.Auth, from string, to []string, body []byte) error {
	fmt.Println("")
	fmt.Println(string(body))
	fmt.Println("")
	return nil
}

type EmailViaSMTPConfig struct {
	Username  string
	Password  string
	SelfEmail string
	DoNotSend bool
}

func (c EmailViaSMTPConfig) ToNotifier() *EmailViaSMTPNotifier {
	notifier := NewEmailViaGmailNotifier(c.Username, c.Password, c.SelfEmail)
	if c.DoNotSend {
		notifier.sendMailFunc = printMail
	}
	return notifier
}

func NewEmailViaSMTPNotifier(hostname string, port int, authenticator smtp.Auth, selfEmail string) *EmailViaSMTPNotifier {
	return &EmailViaSMTPNotifier{
		Hostname:      hostname,
		Port:          port,
		Authenticator: authenticator,
		SelfEmailAddr: selfEmail,
		sendMailFunc:  smtp.SendMail,
	}
}

func NewEmailViaGmailNotifier(username string, password string, selfEmail string) *EmailViaSMTPNotifier {
	return &EmailViaSMTPNotifier{
		Hostname:      "smtp.gmail.com",
		Port:          587,
		Authenticator: smtp.PlainAuth("", username, password, "smtp.gmail.com"),
		SelfEmailAddr: selfEmail,
		sendMailFunc:  smtp.SendMail,
	}
}

func (n *EmailViaSMTPNotifier) Name() string {
	return "EmailViaSMTP"
}

func (n *EmailViaSMTPNotifier) Send(notifications []Notification, subscriber Subscriber) error {
	if len(notifications) == 1 {
		return n.SendOne(notifications[0], subscriber)
	} else if len(notifications) > 1 {
		return n.SendMany(notifications, subscriber)
	}

	// We don't care about the case when no notification is sent.
	return nil
}

func (n *EmailViaSMTPNotifier) convertTextToCRLF(text string) string {
	text = strings.TrimSpace(text)
	lines := newlineRegex.Split(text, -1)
	return strings.Join(lines, "\r\n")
}

func (n *EmailViaSMTPNotifier) addr() string {
	return fmt.Sprintf("%s:%d", n.Hostname, n.Port)
}

func (n *EmailViaSMTPNotifier) SendOne(notification Notification, subscriber Subscriber) error {
	nData := newNotificationData(notification)
	data := emailTemplateData{
		From:    n.SelfEmailAddr,
		To:      subscriber.Email,
		Subject: fmt.Sprintf("[%s][%s] %s", notification.Channel, notification.Origin, notification.Subject),
		Body:    n.convertTextToCRLF(fmt.Sprintf("%s\r\n\r\nCreated At: %s", nData.Content, nData.Timestamp)),
	}
	return n.sendMail(data)
}

func (n *EmailViaSMTPNotifier) SendMany(notifications []Notification, subscriber Subscriber) error {
	// We assume all notifications comes with the same channel
	data := emailTemplateData{
		From:    n.SelfEmailAddr,
		To:      subscriber.Email,
		Subject: fmt.Sprintf("[%s] Received %d notifications", notifications[0].Channel, len(notifications)),
	}

	emailBodyBuf := &bytes.Buffer{}
	for _, notification := range notifications {
		err := singleNotificationTemplate.Execute(emailBodyBuf, newNotificationData(notification))
		if err != nil {
			return err
		}
		// We want an extra line in this email.
		_, err = emailBodyBuf.WriteString("\r\n")
		if err != nil {
			return err
		}
	}

	data.Body = n.convertTextToCRLF(emailBodyBuf.String())
	return n.sendMail(data)
}

func (n *EmailViaSMTPNotifier) sendMail(data emailTemplateData) error {
	emailMessage, err := data.emailBytes()
	if err != nil {
		return err
	}

	return n.sendMailFunc(n.addr(), n.Authenticator, n.SelfEmailAddr, []string{data.To}, emailMessage)
}
