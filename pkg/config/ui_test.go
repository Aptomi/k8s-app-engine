package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigUI(t *testing.T) {
	config := &UI{
		Schema: "http",
		Host:   "127.0.0.1",
		Port:   12345,
	}
	assert.Equal(t, "http://127.0.0.1:12345", config.URL(), "URL must be correct for UI config")
	assert.Equal(t, "127.0.0.1:12345", config.ListenAddr(), "ListenAddr must be correct for UI config")
}
