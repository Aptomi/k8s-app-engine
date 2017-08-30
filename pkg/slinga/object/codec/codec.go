package codec

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type MarshalUnmarshaler interface {
	GetName() string
	MarshalOne(object object.Base) ([]byte, error)
	MarshalMany(objects []object.Base) ([]byte, error)
	UnmarshalOne(data []byte) (object.Base, error)
	UnmarshalOneOrMany(data []byte) ([]object.Base, error)
}
