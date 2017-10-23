package config

import "fmt"

// API represents configs for the API used in both client and server
type API struct {
	Schema    string `valid:"required"`
	Host      string `valid:"host,required"`
	Port      string `valid:"port,required"`
	APIPrefix string `valid:"required"`
}

// URL returns server API url to connect to
func (a API) URL() string {
	return fmt.Sprintf("%s://%s:%s/%s", a.Schema, a.Host, a.Port, a.APIPrefix)
}

// ListenAddr returns address server listens on
func (a API) ListenAddr() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}
