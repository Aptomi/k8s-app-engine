package util

import (
	"fmt"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"os"
	"path/filepath"
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

// FindYamlFiles returns all files found for given list of file paths that could be of the following types:
// * specific file path
// * directory (then all *.yaml files will be taken from it and all subdirectories)
// * file pattern like
func FindYamlFiles(filePaths []string) ([]string, error) {
	allFiles := make([]string, 0, len(filePaths))

	for _, rawPolicyPath := range filePaths {
		policyPath, errPath := filepath.Abs(rawPolicyPath)
		if errPath != nil {
			return nil, fmt.Errorf("error reading filepath: %s", errPath)
		}

		// if it's a directory, use all yaml files from it
		if stat, err := os.Stat(policyPath); err == nil && stat.IsDir() {
			// if dir provided, use all yaml files from it
			files, errGlob := zglob.Glob(filepath.Join(policyPath, "**", "*.yaml"))
			if errGlob != nil {
				return nil, fmt.Errorf("error while searching yaml files in '%s' (error: %s)", policyPath, err)
			}
			allFiles = append(allFiles, files...)
			continue
		}

		// otherwise, try as a single file or glob pattern/mask (so we can feed wildcard mask and process multiple files)
		files, errGlob := zglob.Glob(policyPath)
		if errGlob != nil {
			return nil, fmt.Errorf("error while searching yaml files in '%s' (error: %s)", policyPath, errGlob)
		}
		if len(files) > 0 {
			allFiles = append(allFiles, files...)
			continue
		}

		return nil, fmt.Errorf("path doesn't exist or no YAML files found under: %s", policyPath)
	}

	return allFiles, nil
}
