package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigAPI(t *testing.T) {
	config := &API{
		Schema:    "http",
		Host:      "127.0.0.1",
		Port:      12345,
		APIPrefix: "v10",
	}
	assert.Equal(t, "http://127.0.0.1:12345/v10", config.URL(), "URL must be correct for API config")
	assert.Equal(t, "127.0.0.1:12345", config.ListenAddr(), "ListenAddr must be correct for API config")
}
