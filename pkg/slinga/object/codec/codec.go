package codec

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

// MarshallerUnmarshaller allows to marshal and unmarshal base objects into a set of bytes
type MarshallerUnmarshaller interface {
	GetName() string
	MarshalOne(object object.Base) ([]byte, error)
	MarshalMany(objects []object.Base) ([]byte, error)
	UnmarshalOne(data []byte) (object.Base, error)
	UnmarshalOneOrMany(data []byte) ([]object.Base, error)
}
