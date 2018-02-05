package config

import (
	"time"
)

// Client is the aptomictl config representation
type Client struct {
	Debug  bool       `yaml:",omitempty" validate:"-"`
	Output string     `yaml:",omitempty" validate:"required"`
	API    API        `yaml:",omitempty" validate:"required"`
	Auth   ClientAuth `yaml:",omitempty" validate:"required"`
	HTTP   HTTP       `yaml:",omitempty" validate:"required"`
}

// HTTP is the config for low level HTTP client
type HTTP struct {
	Timeout time.Duration `yaml:",omitempty" `
}

// IsDebug returns true if debug mode enabled
func (c Client) IsDebug() bool {
	return c.Debug
}

// ClientAuth represents client auth configs
type ClientAuth struct {
	Token string `yaml:",omitempty" validate:"-"`
}
