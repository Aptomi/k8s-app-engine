package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigServer(t *testing.T) {
	config := &Server{}
	assert.Equal(t, false, config.IsDebug(), "IsDebug() must be false for default server config")
}
