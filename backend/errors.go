package backend

import "fmt"

type ChannelNotFound struct {
	ChannelName string
}

func (err ChannelNotFound) Error() string {
	return fmt.Sprintf("channel '%s' not found", err.ChannelName)
}
