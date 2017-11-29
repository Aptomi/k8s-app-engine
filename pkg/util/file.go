package util

import (
	"io/ioutil"
)

// WriteTempFile creates a temporary file, writes given data into it and returns its name.
// It's up to a caller to delete the created temporary file by calling os.Remove() on its name.
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
