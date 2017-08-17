package db2

import (
	"fmt"
	"github.com/pborman/uuid"
	"strconv"
	"strings"
	"sync"
)

// Generation represents object's "version" and starts from 0
type Generation uint64

// UID represents unique object ID.
// It's needed because object with some name could be removed and than created again and in this case generation should
// start from 0 again. In this case, on new object creation new UID always created that guarantees differentiation
// between objects created with same name and deleted.
type UID string

// It helps to make sure that UID is unique
var (
	uuidCreationLock sync.Mutex
	lastCreatedUUID  uuid.UUID
)

// NewUUID creates new guaranteed unique thread-safe unique ID
func NewUUID() UID {
	uuidCreationLock.Lock()
	defer uuidCreationLock.Unlock()

	newUUID := uuid.NewUUID()

	// Identical UUIDs could be generated in case of small time interval
	// between NewUUID calls. Let's wait until new UUID generated.
	// UUID uses 100 ns increments, so, it's okay to just poll for new value
	for uuid.Equal(lastCreatedUUID, newUUID) == true {
		// TODO(slukjanov): replace with some WARN message and/or event
		fmt.Println("Same UUID generated!!!")
		newUUID = uuid.NewUUID()
	}
	lastCreatedUUID = newUUID

	return UID(newUUID.String())
}

// KeySeparator used to separate UID and Generation inside the Key
const KeySeparator = "$"

// Key represents unified object's key that includes object's UID and generation.
// So, it means that Key could be used to reference concrete object with concrete
// generation (while UID is a reference to the concrete object but any generation).
type Key string

func (key Key) parts() []string {
	parts := strings.Split(string(key), "$")
	if len(parts) != 2 {
		panic("Key should consist of two parts separated by " + KeySeparator)
	}
	return parts
}

// GetUID returns UID part of the Key
func (key Key) GetUID() UID {
	return UID(key.parts()[0])
}

// GetGeneration returns Generation part of the Key
func (key Key) GetGeneration() Generation {
	val, err := strconv.ParseUint(key.parts()[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return Generation(val)
}
