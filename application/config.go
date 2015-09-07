package main

import (
	"encoding/json"
	"os"

	"gitlab.com/shuhao/towncrier/webreceiver"
)

type ApplicationConfig struct {
	BackendName       string
	BackendOpenString string

	Receiver webreceiver.ReceiverConfig
	// Feed webfeed.FeedConfig
}

func NewApplicationConfig(path string) (*ApplicationConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := &ApplicationConfig{}
	err = json.NewDecoder(f).Decode(c)
	return c, err
}
