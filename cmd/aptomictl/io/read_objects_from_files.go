package io

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/util"
	log "github.com/sirupsen/logrus"
	yamlv2 "gopkg.in/yaml.v2"
)

// ReadLangObjects scans the provided files/dirs/stdin, finds Aptomi lang objects, parses and returns them
func ReadLangObjects(policyPaths []string) ([]runtime.Object, error) {
	policyReg := runtime.NewRegistry().Append(lang.PolicyObjects...)
	codec := yaml.NewCodec(policyReg)

	if len(policyPaths) == 1 && policyPaths[0] == "-" {
		return readLangObjectsFromStdin(codec)
	} else if len(policyPaths) > 0 {
		return readLangObjectsFromFiles(policyPaths, codec)
	}

	return nil, fmt.Errorf("policy file path is not specified")
}

func readLangObjectsFromStdin(codec runtime.Codec) ([]runtime.Object, error) {
	log.Info("Applying policy from stdin")
	data, readErr := ioutil.ReadAll(os.Stdin)
	if readErr != nil {
		return nil, fmt.Errorf("error while reading from stdin")
	}

	objects, decodeErr := codec.DecodeOneOrMany(data)
	if decodeErr != nil {
		return nil, fmt.Errorf("can't unmarshal stdin: %s", decodeErr)
	}

	for _, obj := range objects {
		if !lang.IsPolicyObject(obj) {
			return nil, fmt.Errorf("only policy objects could be applied but got: %s", obj.GetKind())
		}

		if _, ok := obj.(lang.Base); !ok {
			return nil, fmt.Errorf("only policy objects could be applied but got: %s (can't cast to lang.Base)", obj.GetKind())
		}
	}

	return objects, nil
}

func readLangObjectsFromFiles(policyPaths []string, codec runtime.Codec) ([]runtime.Object, error) {
	files, err := findPolicyFiles(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("error while searching for policy files: %s", err)
	}

	log.Info("Loading policy objects:")

	allObjects := make([]runtime.Object, 0)
	objectFile := make(map[string]string)

FILES:
	for _, file := range files {
		data, readErr := ioutil.ReadFile(file)
		if readErr != nil {
			return nil, fmt.Errorf("can't read file %s error: %s", file, readErr)
		}

		// skip entire file if we think that it's a file with k8s objects
		if isK8sObject(data) {
			continue FILES
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

			if service, serviceOk := obj.(*lang.Service); serviceOk {
				for _, component := range service.Components {
					if component.Code == nil || component.Code.Params == nil {
						continue
					}

					includeErr := util.ProcessIncludeMacros(component.Code.Params, filepath.Dir(file))
					if includeErr != nil {
						return nil, includeErr
					}
				}
			}
		}

		log.Infof("  [*] %s", file)

		for _, obj := range objects {
			langObj := obj.(lang.Base) // nolint: errcheck
			log.Infof("\t -> %s %s in %s", langObj.GetKind(), langObj.GetName(), langObj.GetNamespace())
		}

		allObjects = append(allObjects, objects...)
	}

	if len(allObjects) == 0 {
		return nil, fmt.Errorf("no objects found in %s", policyPaths)
	}

	return allObjects, nil
}

func findPolicyFiles(policyPaths []string) ([]string, error) {
	allFiles, err := util.FindYamlFiles(policyPaths)
	if err != nil {
		return nil, err
	}

	sort.Strings(allFiles)

	return allFiles, nil
}

func isK8sObject(data []byte) bool {
	k8sObj := make(map[string]interface{})
	k8sErr := yamlv2.Unmarshal(data, k8sObj)
	if k8sErr != nil {
		return false
	}

	// Aptomi don't have a spec field in language objects, but k8s always have it in all objects
	_, exist := k8sObj["spec"]

	return exist
}
