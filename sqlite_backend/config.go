package sqlite_backend

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gorhill/cronexpr"
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

func (c *Channel) IsTimeToNotifyValid() bool {
	if c.TimeToNotify == ChannelSendImmediately {
		return true
	}

	_, err := cronexpr.Parse(c.TimeToNotify)
	if err != nil {
		logger.WithField("time_to_notify", c.TimeToNotify).Warn("failed to parse time to notify")
	}
	return err == nil
}

func (c *Channel) ShouldSendImmediately() bool {
	return c.TimeToNotify == ChannelSendImmediately
}

func (c *Channel) ShouldSendNowGivenTime(currentTime time.Time) bool {
	// the additional -1s is to make sure that the range always had 1 second.
	// example: if you specify 01:00:00, the minute before and minute after will
	// be 01:00:00 and 01:01:00. The next time computed will be an hour from now
	// (in the hourly case). This means that there's a chance that this hour we
	// never send anything. With an additional second it minimize this risk.
	//
	// Also, double sending is minimized as only 1 goroutine sends and we use a
	// delivered flag in the db.
	minuteBefore := currentTime.Add(time.Duration(-currentTime.Second()-1) * time.Second)
	minuteAfter := minuteBefore.Add(time.Minute + time.Second)
	expression := cronexpr.MustParse(c.TimeToNotify)
	nextTime := expression.Next(minuteBefore)
	if nextTime.IsZero() {
		return false
	}

	// Operator overloading is so good.
	// return (minuteBefore <= nextTime <= currentTime)
	return (nextTime.After(minuteBefore) || nextTime.Equal(minuteBefore)) && (nextTime.Before(minuteAfter) || nextTime.Equal(minuteAfter))
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
		if !channel.IsTimeToNotifyValid() {
			return fmt.Errorf("channel '%s' has an invalid TimeToNotify", channel.Name)
		}
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
