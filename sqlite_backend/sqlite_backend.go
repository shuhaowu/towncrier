package sqlite_backend

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/shuhao/towncrier/backend"
	"gopkg.in/gorp.v1"
)

const (
	BackendName = "sqlite"
)

type SQLiteNotificationBackend struct {
	*gorp.DbMap

	// Set this if you are embedding this struct directly or via a variable
	// as you just want to use the web to show a dashboard and want to use
	// maybe another way to send notifications.
	NeverSendNotifications bool

	config      *Config
	quitChannel chan struct{}

	forceConfigReload         chan struct{}
	forceNotificationDelivery chan struct{}

	// This channel needs information on it n times before the backend is ready
	started chan struct{}
}

var realLogger = logrus.New()
var logger = realLogger.WithField("backend", BackendName)

func init() {
	notificationBackend := &SQLiteNotificationBackend{
		NeverSendNotifications:    false,
		quitChannel:               make(chan struct{}),
		forceConfigReload:         make(chan struct{}),
		forceNotificationDelivery: make(chan struct{}),
		started:                   make(chan struct{}),
	}

	backend.RegisterBackend(notificationBackend)
}

// Initializes a new instance of the backend
//
// This function must be called once only after getting a backend as per the
// NotificationBackend specification.
//
// The openString format is as follows:
//
//     <db_file_path>,<config_file_path>
func (b *SQLiteNotificationBackend) Initialize(openString string) error {
	data := strings.Split(openString, ",")

	db, err := sql.Open("sqlite3", data[0])
	if err != nil {
		return fmt.Errorf("could not open db %s with error: %v", data[0], err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// We need to ignore tags as that's a []string
	table := dbmap.AddTableWithName(Notification{}, "notifications").SetKeys(true, "id")
	table.ColMap("Tags").SetTransient(true)
	table.ColMap("Priority").SetTransient(true)

	config, err := LoadConfig(data[1])
	if err != nil {
		return fmt.Errorf("could not open config at '%s/' with error: %v", data[1], err)
	}

	b.DbMap = dbmap
	b.config = config
	return nil
}

func (b *SQLiteNotificationBackend) QueueNotification(notification backend.Notification) error {
	localNotification := &Notification{
		Notification: notification,
	}

	err := b.saveNotification(localNotification)
	if err != nil {
		return err
	}

	if b.NeverSendNotifications {
		return nil
	}

	return b.conditionallySendNotification(func(channel *Channel, n backend.Notifier) bool {
		return n.ShouldSendImmediately() || channel.ShouldSendImmediately()
	}, localNotification)
}

func (b *SQLiteNotificationBackend) Name() string {
	return "sqlite"
}

func (b *SQLiteNotificationBackend) Start(wg *sync.WaitGroup) {
	wg.Add(2)
	go func() {
		defer wg.Done()

		b.startConfigReloader()
	}()

	go func() {
		defer wg.Done()

		b.startNotificationDelivery()
	}()
}

func (b *SQLiteNotificationBackend) Shutdown() {
	close(b.quitChannel)
	close(b.forceConfigReload)
	close(b.forceNotificationDelivery)

	b.BlockUntilReady()
}

func (b *SQLiteNotificationBackend) BlockUntilReady() {
	for i := 0; i < numberOfBackgroundTasks; i++ {
		_, open := <-b.started
		if !open {
			return
		}
	}

	close(b.started)
}

func (b *SQLiteNotificationBackend) startedOneTask() {
	b.started <- struct{}{}
}

// When using this function in combination with GetSubscribers(), there is no
// guarentee that the two lists are not coming from two different versions of
// the configuration.
//
// There is no public methods for getting them under one version.
func (b *SQLiteNotificationBackend) GetChannels() []*Channel {
	channels := b.config.Channels
	channelsList := make([]*Channel, len(channels))
	i := 0
	for _, channel := range channels {
		channelsList[i] = channel
		i++
	}

	return channelsList
}

func (b *SQLiteNotificationBackend) ForceConfigReload() {
	b.forceConfigReload <- struct{}{}
}

func (b *SQLiteNotificationBackend) ForceNotificationDelivery() {
	b.forceNotificationDelivery <- struct{}{}
}

func (b *SQLiteNotificationBackend) GetSubscribers() []backend.Subscriber {
	subscribers := b.config.Subscribers
	subscribersList := make([]backend.Subscriber, len(subscribers))
	i := 0
	for _, subscriber := range subscribers {
		subscribersList[i] = subscriber
		i++
	}

	return subscribersList
}
