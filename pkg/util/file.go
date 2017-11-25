package util

import (
	"io/ioutil"
)

// WriteTempFile creates a temporary file and returns its name
// todo remove temp files when no more needed
func WriteTempFile(prefix string, data []byte) string {
	tmpFile, err := ioutil.TempFile("", "aptomi-"+prefix)
	if err != nil {
		panic("Failed to create temp file")
	}
	defer tmpFile.Close() // nolint: errcheck

	_, err = tmpFile.Write(data)
	if err != nil {
		panic("Failed to write to temp file")
	}

	return tmpFile.Name()
}
