package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigServer(t *testing.T) {
	config := &Server{}
	assert.Equal(t, false, config.IsDebug(), "IsDebug() must be false for default server config")
}
