package config

import "fmt"

// API represents configs for the API used in both client and server
type API struct {
	Schema    string `yaml:",omitempty" validate:"required"`
	Host      string `yaml:",omitempty" validate:"required,hostname|ip"`
	Port      int    `yaml:",omitempty" validate:"required,min=1,max=65535"`
	APIPrefix string `yaml:",omitempty" validate:"required"`
}

// URL returns server API url to connect to
func (a API) URL() string {
	return fmt.Sprintf("%s://%s:%d/%s", a.Schema, a.Host, a.Port, a.APIPrefix)
}

// ListenAddr returns address server listens on
func (a API) ListenAddr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
