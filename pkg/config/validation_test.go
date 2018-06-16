package config

import (
	"os"
	"testing"

	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Host     string `validate:"required,hostname|ip"`
	Port     int    `validate:"required,min=1,max=65535"`
	SomeDir  string `validate:"required,dir"`
	SomeFile string `validate:"omitempty,file"`
}

func (t *testStruct) IsDebug() bool {
	return false
}

func (t *testStruct) GetLogLevel() logrus.Level {
	if t.IsDebug() {
		return logrus.DebugLevel
	}
	return logrus.InfoLevel
}

func displayErrorMessages() bool {
	return false
}

func TestConfigValidation(t *testing.T) {
	tmpFile := util.WriteTempFile("unittest", []byte("unittest"))
	defer os.Remove(tmpFile) // nolint: errcheck

	tests := []struct {
		config Base
		result bool
	}{
		{
			&testStruct{
				Host:    "0.0.0.0",
				Port:    80,
				SomeDir: "/tmp",
			},
			true,
		},
		{
			&testStruct{
				Host:    "0.0.0.0",
				Port:    80,
				SomeDir: "",
			},
			false,
		},
		{
			&testStruct{
				Host:    "0.0.0.0",
				Port:    80,
				SomeDir: "/nonexistingdirectoryinroot",
			},
			false,
		},
		{
			&testStruct{
				Host:    "127.0.0.1",
				Port:    8080,
				SomeDir: "/tmp",
			},
			true,
		},
		{
			&testStruct{
				Host:    "10.20.30.40",
				Port:    65080,
				SomeDir: "/tmp",
			},
			true,
		},
		{
			&testStruct{
				Host:    "demo.aptomi.io",
				Port:    65080,
				SomeDir: "/tmp",
			},
			true,
		},
		{
			&testStruct{
				Host:    "0.0.0.0",
				Port:    0,
				SomeDir: "/tmp",
			},
			false,
		},
		{
			&testStruct{
				Host:    "0.0.0.0",
				Port:    -1,
				SomeDir: "/tmp",
			},
			false,
		},
		{
			&testStruct{
				Host:    "",
				Port:    80,
				SomeDir: "/tmp",
			},
			false,
		},
		{
			&testStruct{
				Host:     "0.0.0.0",
				Port:     80,
				SomeDir:  "/tmp",
				SomeFile: tmpFile,
			},
			true,
		},
		{
			&testStruct{
				Host:     "0.0.0.0",
				Port:     80,
				SomeDir:  "/tmp",
				SomeFile: tmpFile + ".non-existing",
			},
			false,
		},
	}
	for _, test := range tests {
		val := NewValidator(test.config)
		err := val.Validate()
		failed := !assert.Equal(t, test.result, err == nil, "Validation test case failed: %s", test.config)
		if err != nil {
			msg := err.Error()
			if displayErrorMessages() || failed {
				t.Log(msg)
			}
		}
	}
}
