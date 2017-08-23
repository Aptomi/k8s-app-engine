package object

import (
	"github.com/satori/go.uuid"
	"strconv"
	"strings"
)

// Generation represents object's "version" and starts from 0
type Generation uint64

// String returns generation as string to implement Stringer interface
func (generation Generation) String() string {
	return strconv.FormatUint(uint64(generation), 10)
}

// UID represents unique object ID.
// It's needed because object with some name could be removed and than created again and in this case generation should
// start from 1 again. In this case, on new object creation new UID always created that guarantees differentiation
// between objects created with same name and deleted.
type UID string

// String returns UID as string to implement Stringer interface
func (uid UID) String() string {
	return string(uid)
}

// NewUUID creates new guaranteed unique thread-safe unique ID
func NewUUID() UID {
	return UID(uuid.NewV1().String())
}

// KeySeparator used to separate UID and Generation inside the Key
const KeySeparator = "$"

// Key represents unified object's key that includes object's UID and generation.
// So, it means that Key could be used to reference concrete object with concrete
// generation (while UID is a reference to the concrete object but any generation).
type Key string

func (key Key) parts() []string {
	parts := strings.Split(string(key), KeySeparator)
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

// KeyFromParts return uid and generation combined into Key
func KeyFromParts(uid UID, generation Generation) Key {
	return Key(uid.String() + KeySeparator + generation.String())
}
