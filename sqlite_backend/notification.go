package sqlite_backend

import (
	"strings"

	"github.com/Sirupsen/logrus"

	"gitlab.com/shuhao/towncrier/backend"
	"gopkg.in/gorp.v1"
)

type Notification struct {
	Id          int64 `db:"id"`
	TagsString  string
	PriorityInt int64
	Delivered   bool
	backend.Notification
}

func (n *Notification) predatabaseOp() {
	n.TagsString = strings.Join(n.Tags, ",")
	n.PriorityInt = int64(n.Priority)
}

func (n *Notification) PreInsert(s gorp.SqlExecutor) error {
	n.predatabaseOp()
	return nil
}

func (n *Notification) PreUpdate(s gorp.SqlExecutor) error {
	n.predatabaseOp()
	return nil
}

func (n *Notification) PostGet(s gorp.SqlExecutor) error {
	n.Tags = strings.Split(n.TagsString, ",")
	n.Priority = backend.Priority(n.PriorityInt)
	return nil
}

func (n *Notification) save(dbmap *gorp.DbMap) error {
	return dbmap.Insert(n)
}

func (n *Notification) send(dbmap *gorp.DbMap, channel *Channel, subscribers []backend.Subscriber) error {
	failedToSendError := NewNotificationFailedToSendToSubscribersError(n)

	for _, subscriber := range subscribers {
		for _, notifierName := range channel.Notifiers {
			notifier := backend.GetNotifier(notifierName)
			if notifier == nil {
				logger.WithField("notifier", notifierName).Warnf("cannot find notifier")
				continue
			}

			err := notifier.Send(n.Notification, subscriber)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error":      err,
					"subscriber": subscriber.UniqueName,
					"notifier":   notifierName,
				}).Errorf("failed to send notification")
				failedToSendError.AddError(subscriber.UniqueName, err)
			}
		}
	}

	// Will error if anything errors...
	//
	// !!!  WARNING  !!!
	// !!! YOLOSCALE !!!
	// !!!  WARNING  !!!
	//
	// Not scalable. Need to only error if all failed, but that requires
	// architectural changes from the error stuff.
	// TODO: fix.
	if failedToSendError.HasError() {
		return failedToSendError
	}

	n.Delivered = true
	_, err := dbmap.Update(n)
	if err != nil {
		return err
	}

	return nil
}

func (b *SQLiteNotificationBackend) sendNotificationsIfTimeIsUp() error {
	return nil
}
