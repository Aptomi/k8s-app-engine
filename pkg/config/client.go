package config

import "fmt"

type Client struct {
	Debug  bool
	Server clientServer
	Apply  Apply
}

func (c Client) IsDebug() bool {
	return c.Debug
}

// todo rename to API and share with Server
type clientServer struct {
	Host string `valid:"host,required"`
	Port string `valid:"port,required"`
	// api prefix
}

func (s clientServer) URL() string {
	return fmt.Sprintf("http://%s:%d/api/v1/policy", s.Host, s.Port)
}

type Apply struct {
	PolicyPaths []string `valid:"required"`
}
