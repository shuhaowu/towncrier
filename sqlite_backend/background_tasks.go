package sqlite_backend

import "time"

const (
	reloadConfigInterval = 5 * time.Minute
)

func (b *SQLiteNotificationBackend) startConfigReloader() {
	for {
		select {
		case <-b.quitChannel:
			logger.Info("shutting down config reloader")
			return
		case <-time.After(reloadConfigInterval):
			logger.Info("reloading config")
			err := b.config.Reload()
			if err != nil {
				logger.WithField("error", err).Error("failed to reload configuration")
			}
		}
	}
}

func (b *SQLiteNotificationBackend) startNotificationDelivery() {
	for {
		select {
		case <-b.quitChannel:
			logger.Info("shutting down notification delivery")
			return
		}
	}
}
