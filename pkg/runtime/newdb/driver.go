package newdb

import (
	"fmt"
	"sync"
)

type Driver interface {
	Name() string
	Config() Config
	Store(cfg Config) (Store, error)
}

type Config interface {
	fmt.Stringer
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

	name := driver.Name()
	if len(name) == 0 {
		panic(fmt.Sprintf("db: can't register driver with empty name: %T", driver))
	}

	if _, duplicated := drivers[name]; duplicated {
		panic("db: register called twice for driver: " + name)
	}

	drivers[name] = driver
}

func GetDriverConfig(name string) (Config, error) {
	driversMu.RLock()
	defer driversMu.RUnlock()

	if driver, exists := drivers[name]; exists {
		return driver.Config(), nil
	}

	return nil, fmt.Errorf("can't find driver to get config: %s", name)
}

func New(name string, cfg Config) (Store, error) {
	driversMu.RLock()
	defer driversMu.RUnlock()

	if driver, exists := drivers[name]; exists {
		return driver.Store(cfg)
	}

	return nil, fmt.Errorf("can't find driver to create store: %s", name)
}
