package util

// based on k8s uuid usage

import (
	"fmt"
	"github.com/pborman/uuid"
	"sync"
)

// UID represents unique ID
type UID string

var uuidCreationLock sync.Mutex
var lastCreatedUUID uuid.UUID

// NewUUID creates new guaranteed unique thread-safe unique ID
func NewUUID() UID {
	uuidCreationLock.Lock()
	defer uuidCreationLock.Unlock()

	newUUID := uuid.NewUUID()

	// Identical UUIDs could be generated in case of small time interval
	// between NewUUID calls. Let's wait until new UUID generated.
	// UUID uses 100 ns increments, so, it's okay to just poll for new value
	for uuid.Equal(lastCreatedUUID, newUUID) == true {
		fmt.Println("Same UUID generated!!!")
		newUUID = uuid.NewUUID()
	}
	lastCreatedUUID = newUUID

	return UID(newUUID.String())
}
