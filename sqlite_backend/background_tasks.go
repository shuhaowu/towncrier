package sqlite_backend

import "time"

const (
	reloadConfigInterval         = 5 * time.Minute
	notificationDeliveryInterval = time.Minute
)

func (b *SQLiteNotificationBackend) doConfigReloadLogIfError() {
	logger.Info("reloading config")
	err := b.config.Reload()
	if err != nil {
		logger.WithField("error", err).Error("failed to reload configuration")
	}
}

func (b *SQLiteNotificationBackend) deliverNotificationLogIfError() {

}

func (b *SQLiteNotificationBackend) startConfigReloader() {
	for {
		select {
		case <-b.quitChannel:
			logger.Info("shutting down config reloader")
			return
		case <-time.After(reloadConfigInterval):
			b.doConfigReloadLogIfError()
		case <-b.forceConfigReload:
			b.doConfigReloadLogIfError()
		}
	}
}

func (b *SQLiteNotificationBackend) startNotificationDelivery() {
	for {
		select {
		case <-b.quitChannel:
			logger.Info("shutting down notification delivery")
			return
		case <-time.After(notificationDeliveryInterval):
			b.deliverNotificationLogIfError()
		case <-b.forceNotificationDelivery:
			b.deliverNotificationLogIfError()
		}
	}
}
