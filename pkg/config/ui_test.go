package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigUI(t *testing.T) {
	config := &UI{
		Enable: true,
	}
	assert.Equal(t, true, config.Enable, "UI should be enabled by default")
}
