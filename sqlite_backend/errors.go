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
	Errors        map[string]error // subscriber unique name to error
	Notifications []*Notification
}

func NewNotificationFailedToSendToSubscribersError(notifications []*Notification) *NotificationFailedtoSendToSomeSubscribers {
	return &NotificationFailedtoSendToSomeSubscribers{
		Errors:        make(map[string]error),
		Notifications: notifications,
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
	idsStringArray := make([]string, len(err.Notifications))

	for i, n := range err.Notifications {
		idsStringArray[i] = fmt.Sprintf("%d", n.Id)
	}

	ids := strings.Join(idsStringArray, ",")

	return fmt.Sprintf("failed to send notifications %s: %s", ids, msg)
}
