package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigClient(t *testing.T) {
	config := &Client{}
	assert.Equal(t, false, config.IsDebug(), "IsDebug() must be false for default client config")
}
