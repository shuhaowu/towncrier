package backend

type Subscriber struct {
	UniqueName  string
	Name        string
	Email       string
	PhoneNumber string
}

type Notifier interface {
	Name() string
	ShouldSendImmediately() bool
	Send(notification Notification, subscriber *Subscriber) error
}

var availableNotifiers map[string]Notifier = make(map[string]Notifier)

func registerNotifier(n Notifier) {
	availableNotifiers[n.Name()] = n
}

func GetNotifier(name string) Notifier {
	notifier, _ := availableNotifiers[name]
	return notifier
}