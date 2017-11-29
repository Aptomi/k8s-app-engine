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

func readFiles(policyPaths []string) ([]runtime.Object, error) {
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

		if stat, err := os.Stat(policyPath); err == nil {
			if stat.IsDir() { // if dir provided, use all yaml files from it
				files, errGlob := zglob.Glob(filepath.Join(policyPath, "**", "*.yaml"))
				if errGlob != nil {
					return nil, fmt.Errorf("error while searching yaml files in directory: %s error: %s", policyPath, err)
				}
				allFiles = append(allFiles, files...)
			} else { // if specific file provided, use it
				allFiles = append(allFiles, policyPath)
			}
		} else if os.IsNotExist(err) {
			return nil, fmt.Errorf("path doesn't exist: %s error: %s", policyPath, err)
		} else {
			return nil, fmt.Errorf("error while processing path: %s", err)
		}
	}

	sort.Strings(allFiles)

	// todo(slukjanov): log list of files from which we're applying policy
	//fmt.Println("Apply policy from following files:")
	//for idx, policyPath := range allFiles {
	//	fmt.Println(idx, "-", policyPath)
	//}

	return allFiles, nil
}
