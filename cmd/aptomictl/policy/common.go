package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

func readLangFromFiles(policyPaths []string) ([]runtime.Object, error) {
	if len(policyPaths) <= 0 {
		return nil, fmt.Errorf("policy file path is not specified")
	}

	policyReg := runtime.NewRegistry().Append(lang.PolicyObjects...)
	codec := yaml.NewCodec(policyReg)

	files, err := findPolicyFiles(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("error while searching for policy files: %s", err)
	}

	allObjects := make([]runtime.Object, 0)
	objectFile := make(map[string]string)
	for _, file := range files {
		data, readErr := ioutil.ReadFile(file)
		if readErr != nil {
			return nil, fmt.Errorf("can't read file %s error: %s", file, readErr)
		}

		objects, decodeErr := codec.DecodeOneOrMany(data)
		if decodeErr != nil {
			return nil, fmt.Errorf("can't unmarshal file %s error: %s", file, decodeErr)
		}

		for _, obj := range objects {
			if !lang.IsPolicyObject(obj) {
				return nil, fmt.Errorf("only policy objects could be applied but got: %s", obj.GetKind())
			}

			langObj, ok := obj.(lang.Base)
			if !ok {
				return nil, fmt.Errorf("only policy objects could be applied but got: %s (can't cast to lang.Base)", obj.GetKind())
			}

			key := runtime.KeyForStorable(langObj)
			if firstFile := objectFile[key]; len(firstFile) > 0 {
				return nil, fmt.Errorf("duplicate object with key %s detected in file %s (first occurrence is in file %s)", key, file, firstFile)
			}
			objectFile[key] = file
		}

		allObjects = append(allObjects, objects...)
	}

	if len(allObjects) == 0 {
		return nil, fmt.Errorf("no objects found in %s", policyPaths)
	}

	return allObjects, nil
}

func findPolicyFiles(policyPaths []string) ([]string, error) {
	allFiles := make([]string, 0, len(policyPaths))

	for _, rawPolicyPath := range policyPaths {
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

		return nil, fmt.Errorf("path doesn't exist or no YAML files found under: %s")
	}

	sort.Strings(allFiles)

	fmt.Println("Applying policy from:")
	for _, policyPath := range allFiles {
		fmt.Println("  [*] " + policyPath)
	}

	return allFiles, nil
}
