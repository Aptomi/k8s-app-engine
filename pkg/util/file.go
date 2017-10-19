package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// CopyFile copies the contents of the file named src to the file named
// by dst. Dst will be overwritten if already exists. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close() // nolint: errcheck
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		errClose := out.Close()
		if err == nil {
			err = errClose
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// DeleteFile deletes a file
func DeleteFile(src string) (err error) {
	return os.Remove(src)
}

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
