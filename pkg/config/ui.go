package config

import "fmt"

// UI represents configs for the UI used in both client and server
type UI struct {
	Schema string `validate:"required"`
	Host   string `validate:"required,hostname|ip"`
	Port   int    `validate:"required,min=1,max=65535"`
}

// URL returns server API url to connect to
func (u UI) URL() string {
	return fmt.Sprintf("%s://%s:%d", u.Schema, u.Host, u.Port)
}

// ListenAddr returns address server listens on
func (u UI) ListenAddr() string {
	return fmt.Sprintf("%s:%d", u.Host, u.Port)
}
