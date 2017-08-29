package client

import (
	"fmt"
	"github.com/mattn/go-zglob"
	"os"
	"path/filepath"
	"sort"
)

func Apply(policyPaths []string) error {
	allFiles := make([]string, 0, len(policyPaths))

	for _, policyPath := range policyPaths {
		policyPath, err := filepath.Abs(policyPath)
		if err != nil {
			panic(fmt.Sprintf("Error reading filepath: %s", err))
		}

		if stat, err := os.Stat(policyPath); err == nil {
			if stat.IsDir() { // if dir provided, use all yaml files from it
				files, err := zglob.Glob(filepath.Join(policyPath, "**", "*.yaml"))
				if err != nil {
					return fmt.Errorf("Error while searching yaml files in directory: %s error: %s", policyPath, err)
				}
				allFiles = append(allFiles, files...)
			} else { // if specific file provided, use it
				allFiles = append(allFiles, policyPath)
			}
		} else if os.IsNotExist(err) {
			panic(fmt.Sprintf("Path doesn't exists: %s error: %s", policyPath, err))
		} else {
			panic(fmt.Sprintf("Error while processing path: %s", err))
		}
	}

	sort.Strings(allFiles) // todo(slukjanov): do we really need to sort files?

	fmt.Println("Apply policy from following files:")
	for idx, policyPath := range allFiles {
		fmt.Println(idx, "-", policyPath)
	}

	return nil
}
