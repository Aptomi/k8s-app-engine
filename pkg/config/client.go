package config

import (
	"time"
)

// Client is the aptomictl config representation
type Client struct {
	Debug  bool       `validate:"-"`
	Output string     `validate:"required"`
	API    API        `validate:"required"`
	Auth   ClientAuth `validate:"required"`
	HTTP   HTTP       `validate:"required"`
}

// HTTP is the config for low level HTTP client
type HTTP struct {
	Timeout time.Duration
}

// IsDebug returns true if debug mode enabled
func (c Client) IsDebug() bool {
	return c.Debug
}

// ClientAuth represents client auth configs
type ClientAuth struct {
	Username string `validate:"-"`
	Password string `validate:"-"`
	Token    string `validate:"-"`
}
