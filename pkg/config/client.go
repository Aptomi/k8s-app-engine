package config

import (
	"time"

	"github.com/sirupsen/logrus"
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

// GetLogLevel returns log level
func (c *Client) GetLogLevel() logrus.Level {
	if c.IsDebug() {
		return logrus.DebugLevel
	}
	return logrus.InfoLevel
}

// ClientAuth represents client auth configs
type ClientAuth struct {
	Token string `yaml:",omitempty" validate:"-"`
}
