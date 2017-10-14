package config

import "fmt"

type API struct {
	Schema    string `valid:"required"`
	Host      string `valid:"host,required"`
	Port      string `valid:"port,required"`
	APIPrefix string `valid:"required"`
}

func (a API) URL() string {
	return fmt.Sprintf("%s://%s:%s/%s/policy", a.Schema, a.Host, a.Port, a.APIPrefix)
}

func (a API) ListenAddr() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}
