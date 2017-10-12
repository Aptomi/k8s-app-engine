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
				API:  API{Host: "0.0.0.0", Port: "80"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			true,
		},
		{
			"success-0.0.0.0:80",
			&Server{
				API:  API{Host: "0.0.0.0", Port: "80"},
				Helm: Helm{ChartsDir: ""},
			},
			false,
		},
		{
			"success-0.0.0.0:80",
			&Server{
				API:  API{Host: "0.0.0.0", Port: "80"},
				Helm: Helm{ChartsDir: "/nonexistingdirectoryinroot"},
			},
			false,
		},
		{
			"success-127.0.0.1:8080",
			&Server{
				API:  API{Host: "127.0.0.1", Port: "8080"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			true,
		},
		{
			"success-10.20.30.40:65080",
			&Server{
				API:  API{Host: "10.20.30.40", Port: "65080"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			true,
		},
		{
			"success-demo.aptomi.io:65080",
			&Server{
				API:  API{Host: "demo.aptomi.io", Port: "65080"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			true,
		},
		{
			"fail-0.0.0.0:0",
			&Server{
				API:  API{Host: "0.0.0.0", Port: "0"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			false,
		},
		{
			"fail-0.0.0.0:-1",
			&Server{
				API:  API{Host: "0.0.0.0", Port: "-1"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			false,
		},
		{
			"fail-:80",
			&Server{
				API:  API{Host: "", Port: "80"},
				Helm: Helm{ChartsDir: "/tmp"},
			},
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Validate(test.config)
			assert.Equal(t, test.result, result)
			if test.result && !result {
				t.Logf("Unexpected validation error: %s", err)
			}
		})
	}
}
