package sqlite_backend

import (
	"fmt"
	"strings"
)

type ChannelNotFound struct {
	ChannelName string
}

func (err ChannelNotFound) Error() string {
	return fmt.Sprintf("channel '%s' not found", err.ChannelName)
}

type NotificationFailedtoSendToSomeSubscribers struct {
	Errors       map[string]error // subscriber unique name to error
	Notification *Notification
}

func NewNotificationFailedToSendToSubscribersError(n *Notification) *NotificationFailedtoSendToSomeSubscribers {
	return &NotificationFailedtoSendToSomeSubscribers{
		Errors:       make(map[string]error),
		Notification: n,
	}
}

func (err *NotificationFailedtoSendToSomeSubscribers) AddError(name string, e error) {
	err.Errors[name] = e
}

func (err *NotificationFailedtoSendToSomeSubscribers) HasError() bool {
	return len(err.Errors) > 0
}

func (err *NotificationFailedtoSendToSomeSubscribers) Error() string {
	errormsg := make([]string, len(err.Errors))
	i := 0
	for name, e := range err.Errors {
		errormsg[i] = fmt.Sprintf("%s => %v", name, e)
		i++
	}
	msg := strings.Join(errormsg, "; ")
	return fmt.Sprintf("failed to send notification %d: %s", err.Notification.Id, msg)
}
