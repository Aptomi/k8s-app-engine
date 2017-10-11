package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		config Base
		result bool
	}{
		{
			"success-0.0.0.0:80",
			&Server{
				Server: serverServer{Host: "0.0.0.0", Port: "80"},
			},
			true,
		},
		{
			"success-127.0.0.1:8080",
			&Server{
				Server: serverServer{Host: "127.0.0.1", Port: "8080"},
			},
			true,
		},
		{
			"success-10.20.30.40:65080",
			&Server{
				Server: serverServer{Host: "10.20.30.40", Port: "65080"},
			},
			true,
		},
		{
			"success-demo.aptomi.io:65080",
			&Server{
				Server: serverServer{Host: "demo.aptomi.io", Port: "65080"},
			},
			true,
		},
		{
			"fail-0.0.0.0:0",
			&Server{
				Server: serverServer{Host: "0.0.0.0", Port: "0"},
			},
			false,
		},
		{
			"fail-0.0.0.0:-1",
			&Server{
				Server: serverServer{Host: "0.0.0.0", Port: "-1"},
			},
			false,
		},
		{
			"fail-:80",
			&Server{
				Server: serverServer{Host: "", Port: "80"},
			},
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, _ := Validate(test.config)
			assert.Equal(t, test.result, result)
		})
	}
}
