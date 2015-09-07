package sqlite_backend

import (
	"strings"
	"time"

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
	if n.Tags == nil || len(n.Tags) == 0 {
		n.TagsString = ""
	} else {
		n.TagsString = strings.Join(n.Tags, ",")
	}

	n.PriorityInt = int64(n.Priority)
}

func (n *Notification) PreInsert(s gorp.SqlExecutor) error {
	n.predatabaseOp()
	n.CreatedAt = time.Now().UnixNano()
	n.UpdatedAt = n.CreatedAt
	return nil
}

func (n *Notification) PreUpdate(s gorp.SqlExecutor) error {
	n.predatabaseOp()
	n.UpdatedAt = time.Now().UnixNano()
	return nil
}

func (n *Notification) PostGet(s gorp.SqlExecutor) error {
	n.Tags = strings.Split(n.TagsString, ",")
	n.Priority = backend.Priority(n.PriorityInt)
	return nil
}

func (n *Notification) insert(dbmap *gorp.DbMap) error {
	return dbmap.Insert(n)
}

func (n *Notification) setDelivered(dbmap *gorp.DbMap) error {
	n.Delivered = true
	_, err := dbmap.Update(n)
	return err
}

func (b *SQLiteNotificationBackend) sendNotifications(notifications []*Notification, channel *Channel, subscribers []backend.Subscriber) error {
	failedToSendError := NewNotificationFailedToSendToSubscribersError(notifications)

	backendNotificationObjects := make([]backend.Notification, len(notifications))

	for i, n := range notifications {
		backendNotificationObjects[i] = n.Notification
	}

	for _, subscriber := range subscribers {
		for _, notifierName := range channel.Notifiers {
			notifier := backend.GetNotifier(notifierName)
			if notifier == nil {
				logger.WithField("notifier", notifierName).Warnf("cannot find notifier")
				continue
			}

			err := notifier.Send(backendNotificationObjects, subscriber)
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

	var err error = nil
	for _, n := range notifications {
		err = n.setDelivered(b.DbMap)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error":        err,
				"notification": n.Id,
			}).Error("failed to set notification to be delivered")
		}
	}

	return err
}
