package config

import "fmt"

type API struct {
	Schema    string `valid:"required"`
	Host      string `valid:"host,required"`
	Port      string `valid:"port,required"`
	ApiPrefix string `valid:"required"`
}

func (a API) URL() string {
	return fmt.Sprintf("%s://%s:%d/%s/policy", a.Schema, a.Host, a.Port, a.ApiPrefix)
}

func (a API) ListenAddr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
