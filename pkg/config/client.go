package config

import (
	"time"
)

// Client is the aptomictl config representation
type Client struct {
	Debug bool `validate:"-"`
	API   API  `validate:"required"`
	Auth  Auth `validate:"required"`
	HTTP  HTTP `validate:"required"`
}

// HTTP is the config for low level HTTP client
type HTTP struct {
	Timeout time.Duration
}

// IsDebug returns true if debug mode enabled
func (c Client) IsDebug() bool {
	return c.Debug
}

// Auth represents client auth configs
type Auth struct {
	Username string `validate:"required"`
}
