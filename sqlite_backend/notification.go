package sqlite_backend

import "gitlab.com/shuhao/towncrier/backend"

type Notification struct {
	Id int64 `db:"id"`
	backend.Notification
}

func (b *SQLiteNotificationBackend) saveNotification(notification *Notification) error {
	return b.Insert(notification)
}

func (b *SQLiteNotificationBackend) conditionallySendNotification(shouldSend func(*Channel, backend.Notifier) bool, notification *Notification) error {
	// We need to lock here because we want to make sure that when we do send
	// a notification, we are not in the middle of a reload for configuration
	// and try to send to the wrong subscriber.
	//
	// Basically, always ensure that we are using one version of the config,
	// not two half copies.
	// Stupid.
	b.config.Lock()
	channel, found := b.config.Channels[notification.Channel]
	subscribers := b.config.Subscribers
	b.config.Unlock()

	if !found {
		return ChannelNotFound{ChannelName: notification.Channel}
	}

	failedToSendError := NewNotificationFailedToSendToSubscribersError(notification)

	for _, subscriberName := range channel.Subscribers {
		subscriber, found := subscribers[subscriberName]
		if !found {
			logger.Warnf("cannot send notification to subscriber %s: not found", subscriberName)
			continue
		}

		for _, notifierName := range channel.Notifiers {
			notifier := backend.GetNotifier(notifierName)
			if notifier == nil {
				logger.Warnf("cannot find notifier %s", notifierName)
				continue
			}

			if shouldSend(channel, notifier) {
				err := notifier.Send(notification.Notification, subscriber)
				if err != nil {
					logger.Errorf("failed to send notification to subscriber '%s' via '%s'", subscriberName, notifierName)
					failedToSendError.AddError(subscriberName, err)
				}
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

	return nil
}

func (b *SQLiteNotificationBackend) sendNotificationsIfTimeIsUp() error {
	return nil
}
