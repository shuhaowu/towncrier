package sqlite_backend

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"gitlab.com/shuhao/towncrier/backend"
)

const (
	subscribersConfigFileName = "subscribers.conf.json"
	channelsConfigFileName    = "channels.conf.json"

	ChannelSendImmediately = "@immediately"
)

type Channel struct {
	Name         string
	Subscribers  []string // a list of unique names as specified by the users field
	Notifiers    []string
	TimeToNotify string
}

func (c *Channel) ShouldSendImmediately() bool {
	return c.TimeToNotify == ChannelSendImmediately
}

func (c *Channel) ShouldSendNowGivenTime(t time.Time) bool {
	return false
}

type ConfigJSON struct {
	Channels    []*Channel
	Subscribers []backend.Subscriber
}

type Config struct {
	// Config has a lock for when it is reloaded, we would like to lock it.
	// Lock should occur when we find channels and its subscribers as well, when
	// we want to send messages.
	*sync.Mutex
	ConfigPath  string
	Channels    map[string]*Channel
	Subscribers map[string]backend.Subscriber
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		Mutex:      &sync.Mutex{},
		ConfigPath: configPath,

		Channels:    make(map[string]*Channel),
		Subscribers: make(map[string]backend.Subscriber),
	}

	return config, config.Reload()
}

func (c *Config) Reload() error {
	file, err := os.Open(c.ConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	configJson := &ConfigJSON{}

	err = json.NewDecoder(file).Decode(configJson)
	if err != nil {
		return err
	}

	channels := make(map[string]*Channel)
	for _, channel := range configJson.Channels {
		channels[channel.Name] = channel
	}

	subscribers := make(map[string]backend.Subscriber)
	for _, subscriber := range configJson.Subscribers {
		subscribers[subscriber.UniqueName] = subscriber
	}

	c.Lock()
	c.Subscribers = subscribers
	c.Channels = channels
	c.Unlock()

	return nil
}
