package bolt

import (
	"fmt"
	"time"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
	"github.com/coreos/bbolt"
)

func init() {
	db.RegisterDriver(&boltDriver{})
}

type boltDriver struct {
}

func (driver *boltDriver) GetName() string {
	return "bolt"
}

func (driver *boltDriver) Open(dataSourceName string) (db.Store, error) {
	boltDB, err := bolt.Open(dataSourceName, 0600, &bolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("error while opening BoltDB: %s error: %s", dataSourceName, err)
	}

	store := &boltStore{
		bolt: boltDB,
	}

	return store, store.init()
}
