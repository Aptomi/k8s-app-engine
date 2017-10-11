package config

import "fmt"

type Server struct {
	Debug  bool
	Server server
}

type server struct {
	Host string `valid:"host,required"`
	Port string `valid:"port,required"`
}

func (s *server) ListenAddr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}
