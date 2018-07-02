package etcd

import (
	"time"
)

var (
	// todo it's an aggressive config to detect failed etcd nodes faster, reconsider
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
	dialTimeout      = 10 * time.Second
)

type Config struct {
	Prefix    string
	Endpoints []string
	// todo add tls config and auth for etcd
}
