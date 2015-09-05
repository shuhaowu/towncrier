package sqlite_backend

import "time"

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

func (b *SQLiteNotificationBackend) deliverNotificationLogIfError() {

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
