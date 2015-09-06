package sqlite_backend

import (
	"time"

	"gitlab.com/shuhao/towncrier/backend"
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
func (b *SQLiteNotificationBackend) deliverNotificationsLogIfError(currentTime time.Time) {
	logger.Info("attempting to deliver notifications")

	// No need to lock as we only access one thing
	channels := b.config.Channels

	for _, channel := range channels {
		if !channel.ShouldSendNowGivenTime(currentTime) {
			continue
		}

		localLog := logger.WithField("channel", channel.Name)

		var notifications []*Notification
		_, err := b.Select(&notifications, "SELECT * FROM notifications WHERE Delivered = 0 AND Channel = ?", channel.Name)
		if err != nil {
			localLog.WithField("error", err).Error("cannot select notifications from the database")
			return
		}

		if len(notifications) == 0 {
			continue
		}

		localLog.Infof("found %d notifications to deliver to %d subscribers", len(notifications), len(channel.Subscribers))

		c, subscribers := b.GetChannelAndItsSubscribers(channel.Name)
		if c == nil {
			localLog.Warnf("channel disappeared during sending, ignoring")
			continue
		}

		// We need this to be parallel as things can block and be really slow.
		go func(notifications []*Notification, channel *Channel, subscribers []backend.Subscriber) {
			err := b.sendNotifications(notifications, channel, subscribers)
			if err != nil {
				localLog.WithField("error", err).Error("failed to send notification")
			}

			localLog.Info("notifications successfully sent")
		}(notifications, c, subscribers)
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

	if b.NeverSendNotifications {
		logger.Info("we should never send notifications, shutting down...")
		goto shutdown
	}

	for {
		select {
		case <-b.quitChannel:
			goto shutdown
		case <-time.After(notificationDeliveryInterval):
			b.deliverNotificationsLogIfError(time.Now())
		case _, open := <-b.forceNotificationDelivery:
			if !open {
				goto shutdown
			}

			b.deliverNotificationsLogIfError(time.Now())
		}
	}

shutdown:
	logger.Info("shutting down notification delivery")
	return
}
