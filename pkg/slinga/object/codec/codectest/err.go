package codectest

import (
	"errors"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
)

type errCodec struct{}

// ErrCodecName is the name of Err MarshalUnmarshaler implementation (always returns err)
const ErrCodecName = "err"

// ErrCodec is an instance of ErrCodec that is fully stateless (and it means thread-safe)
var ErrCodec codec.MarshalUnmarshaler = &errCodec{}

func (c *errCodec) GetName() string {
	return ErrCodecName
}

func (c *errCodec) SetObjectCatalog(catalog *ObjectCatalog) {
	// noop
}
func (c *errCodec) MarshalOne(object BaseObject) ([]byte, error) {
	return nil, errors.New("MarshalOne error")
}

func (c *errCodec) MarshalMany(objects []BaseObject) ([]byte, error) {
	return nil, errors.New("MarshalMany error")
}

func (c *errCodec) UnmarshalOne(data []byte) (BaseObject, error) {
	return nil, errors.New("UnmarshalOne error")
}

func (c *errCodec) UnmarshalOneOrMany(data []byte) ([]BaseObject, error) {
	return nil, errors.New("MarshalOneOrMany error")
}
