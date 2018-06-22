package db

import (
	"fmt"
	"sync"
)

type Driver interface {
	GetName() string
	Open(dataSourceName string) (Store, error)
}

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

func RegisterDriver(driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("db: can't register nil driver")
	}

	name := driver.GetName()
	if len(name) == 0 {
		panic(fmt.Sprintf("db: can't register driver with empty name: %T", driver))
	}

	if _, duplicated := drivers[name]; duplicated {
		panic("db: register called twice for driver: " + name)
	}

	drivers[name] = driver
}

func Open(name string, dataSourceName string) (Store, error) {
	driversMu.RLock()
	defer driversMu.RUnlock()

	if driver, exists := drivers[name]; exists {
		return driver.Open(dataSourceName)
	}

	return nil, fmt.Errorf("can't find driver: %s", name)
}
