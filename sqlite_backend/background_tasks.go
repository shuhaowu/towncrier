package sqlite_backend

import (
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	reloadConfigInterval         = 5 * time.Minute
	notificationDeliveryInterval = time.Minute
	numberOfBackgroundTasks      = 2
)

func (b *SQLiteNotificationBackend) doConfigReloadLogIfError() {
	logger.Info("reloading config")
	err := b.config.Reload()
	if err != nil {
		logger.WithField("error", err).Error("failed to reload configuration")
	}
}

// The algorithm of this function goes as follows:
//
// 1. For each channel, compute if we are at the window of delivery.
// 2. If yes, get the list of notifications that's unsent
// 3. Send each notifications
//
func (b *SQLiteNotificationBackend) deliverNotificationLogIfError() {
	// No need to lock as we only access one thing
	channels := b.config.Channels
	currentTime := time.Now()

	for _, channel := range channels {
		if !channel.ShouldSendNowGivenTime(currentTime) {
			continue
		}

		var notifications []*Notification
		_, err := b.Select(&notifications, "SELECT * FROM notifications WHERE Delivered = 0 AND Channel = ?", channel.Name)
		if err != nil {
			logger.WithField("error", err).Error("cannot select notifications from the database")
			return
		}

		c, subscribers := b.GetChannelAndItsSubscribers(channel.Name)
		if c == nil {
			logger.WithField("channel", channel.Name).Warnf("channel disappeared during sending, ignoring")
			continue
		}

		err = b.sendNotifications(notifications, c, subscribers)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error":   err,
				"channel": c.Name,
			}).Error("failed to send notification")
		}
	}
}

func (b *SQLiteNotificationBackend) startConfigReloader() {
	logger.Info("started config reloader")
	b.startedOneTask()
	for {
		select {
		case <-b.quitChannel:
			goto shutdown
		case <-time.After(reloadConfigInterval):
			b.doConfigReloadLogIfError()
		case _, open := <-b.forceConfigReload:
			if !open {
				goto shutdown
			}

			b.doConfigReloadLogIfError()
		}
	}

shutdown:
	logger.Info("shutting down config reloader")
	return
}

func (b *SQLiteNotificationBackend) startNotificationDelivery() {
	logger.Info("started notification delivery")
	b.startedOneTask()
	for {
		select {
		case <-b.quitChannel:
			goto shutdown
		case <-time.After(notificationDeliveryInterval):
			b.deliverNotificationLogIfError()
		case _, open := <-b.forceNotificationDelivery:
			if !open {
				goto shutdown
			}

			b.deliverNotificationLogIfError()
		}
	}

shutdown:
	logger.Info("shutting down notification delivery")
	return
}
