package backend

type Subscriber struct {
	UniqueName  string
	Name        string
	Email       string
	PhoneNumber string
}

type Notifier interface {
	Name() string
	Send(notification Notification, subscriber Subscriber) error
}

var availableNotifiers map[string]Notifier = make(map[string]Notifier)

func RegisterNotifier(n Notifier) {
	availableNotifiers[n.Name()] = n
}

func UnregisterNotifier(name string) {
	delete(availableNotifiers, name)
}

func ClearAllNotifiers() {
	availableNotifiers = make(map[string]Notifier)
}

func GetNotifier(name string) Notifier {
	notifier, _ := availableNotifiers[name]
	return notifier
}
