package util

import (
	"fmt"
	"io/ioutil"
)

// WriteTempFile creates a temporary file and returns its name
func WriteTempFile(prefix string, content string) string {
	tmpFile, err := ioutil.TempFile("", "aptomi-"+prefix)
	if err != nil {
		panic("Failed to create temp file")
	}
	defer tmpFile.Close() // nolint: errcheck

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		panic("Failed to write to temp file")
	}

	return tmpFile.Name()
}
