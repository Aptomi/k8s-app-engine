package fake

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrCodec(t *testing.T) {
	assert.Equal(t, ErrCodecName, ErrCodec.GetName(), "Correct Name expected")
	err := ErrCodec.Unmarshal(make([]byte, 0), nil)
	assert.NotNil(t, err, "Unmarshal should return error")
	_, err = ErrCodec.Marshal(nil)
	assert.NotNil(t, err, "Marshal should return error")
}
