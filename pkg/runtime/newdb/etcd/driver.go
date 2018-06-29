package etcd

import (
	"fmt"
	"time"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
	etcd "github.com/coreos/etcd/clientv3"
)

func init() {
	newdb.RegisterDriver(&driver{})
}

type config struct {
}

func (c *config) String() string {
	panic("implement me")
}

type driver struct {
}

func (d *driver) Name() string {
	return "etcd"
}

func (d *driver) Config() newdb.Config {
	return &config{}
}

func (d *driver) Store(cfg newdb.Config) (newdb.Store, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("error while creating store using driver %s and config %s: %s", d.Name(), d.Config(), err)
	}

	return &store{client: client}, nil
}
