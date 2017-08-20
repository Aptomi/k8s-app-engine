package codectest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrCodec(t *testing.T) {
	assert.Equal(t, ErrCodecName, ErrCodec.GetName(), "Correct Name expected")

	assert.NotPanics(t, func() { ErrCodec.SetObjectCatalog(nil) }, "")

	_, err := ErrCodec.MarshalOne(nil)
	assert.NotNil(t, err, "MarshalOne should return error")

	_, err = ErrCodec.MarshalMany(nil)
	assert.NotNil(t, err, "MarshalMany should return error")

	_, err = ErrCodec.UnmarshalOne(nil)
	assert.NotNil(t, err, "UnmarshalOne should return error")

	_, err = ErrCodec.UnmarshalOneOrMany(nil)
	assert.NotNil(t, err, "UnmarshalOneOrMany should return error")
}
