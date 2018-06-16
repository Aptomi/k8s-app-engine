package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigUI(t *testing.T) {
	config := &UI{
		Enable: true,
	}
	assert.Equal(t, true, config.Enable, "UI should be enabled by default")
}
