package backend

import (
	"fmt"
	"sync"
)

type NotificationBackend interface {
	Name() string
	QueueNotification(notification Notification) error
	Initialize(openString string) error
	Start(wg *sync.WaitGroup)
	BlockUntilReady()
	Shutdown()
}

var availableBackends map[string]NotificationBackend = make(map[string]NotificationBackend)

func RegisterBackend(backend NotificationBackend) {
	availableBackends[backend.Name()] = backend
}

func GetBackend(name string) NotificationBackend {
	backend, found := availableBackends[name]
	if !found {
		panic(fmt.Sprintf("backend %s not found", name))
	}
	return backend
}

func InitializeNotificationBackend(name, openString string) (NotificationBackend, error) {
	backend := GetBackend(name)
	return backend, backend.Initialize(openString)
}
