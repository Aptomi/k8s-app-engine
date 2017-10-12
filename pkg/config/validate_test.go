package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	Host      string `valid:"host,required"`
	Port      string `valid:"port,required"`
	ChartsDir string `valid:"dir,required"`
}

func (t *testStruct) IsDebug() bool {
	return false
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		config Base
		result bool
	}{
		{
			"success-0.0.0.0:80",
			&testStruct{
				Host:      "0.0.0.0",
				Port:      "80",
				ChartsDir: "/tmp",
			},
			true,
		},
		{
			"success-0.0.0.0:80",
			&testStruct{
				Host:      "0.0.0.0",
				Port:      "80",
				ChartsDir: "",
			},
			false,
		},
		{
			"success-0.0.0.0:80",
			&testStruct{
				Host:      "0.0.0.0",
				Port:      "80",
				ChartsDir: "/nonexistingdirectoryinroot",
			},
			false,
		},
		{
			"success-127.0.0.1:8080",
			&testStruct{
				Host:      "127.0.0.1",
				Port:      "8080",
				ChartsDir: "/tmp",
			},
			true,
		},
		{
			"success-10.20.30.40:65080",
			&testStruct{
				Host:      "10.20.30.40",
				Port:      "65080",
				ChartsDir: "/tmp",
			},
			true,
		},
		{
			"success-demo.aptomi.io:65080",
			&testStruct{
				Host:      "demo.aptomi.io",
				Port:      "65080",
				ChartsDir: "/tmp",
			},
			true,
		},
		{
			"fail-0.0.0.0:0",
			&testStruct{
				Host:      "0.0.0.0",
				Port:      "0",
				ChartsDir: "/tmp",
			},
			false,
		},
		{
			"fail-0.0.0.0:-1",
			&testStruct{
				Host:      "0.0.0.0",
				Port:      "-1",
				ChartsDir: "/tmp",
			},
			false,
		},
		{
			"fail-:80",
			&testStruct{
				Host:      "",
				Port:      "80",
				ChartsDir: "/tmp",
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
