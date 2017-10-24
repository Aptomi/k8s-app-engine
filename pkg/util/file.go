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

// EnsureSingleFile ensures that only one file matches the list of files
func EnsureSingleFile(files []string) (string, error) {
	if len(files) <= 0 {
		return "", fmt.Errorf("no files found")
	}
	if len(files) > 1 {
		return "", fmt.Errorf("more than one file found")
	}
	return files[0], nil
}
