package util

// based on k8s uuid usage

import (
	"github.com/Aptomi/aptomi/pkg/slinga/db2"
	"github.com/pborman/uuid"
	"sync"
)

var uuidCreationLock sync.Mutex
var lastCreatedUUID uuid.UUID

func NewUUID() db2.UID {
	uuidCreationLock.Lock()
	defer uuidCreationLock.Unlock()

	newUUID := uuid.NewUUID()

	// Identical UUIDs could be generated in case of small time interval
	// between NewUUID calls. Let's wait until new UUID generated.
	// UUID uses 100 ns increments, so, it's okay to just poll for new value
	for uuid.Equal(lastCreatedUUID, newUUID) == true {
		newUUID = uuid.NewUUID()
	}
	lastCreatedUUID = newUUID

	return db2.UID(newUUID.String())
}
