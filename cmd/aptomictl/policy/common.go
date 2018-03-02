package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	log "github.com/Sirupsen/logrus"
	yamlv2 "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func readLangObjects(policyPaths []string) ([]runtime.Object, error) {
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
			langObj := obj.(lang.Base)
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

func waitForApplyToFinish(attempts int, interval time.Duration, client client.Core, result *api.PolicyUpdateResult) {
	// if policy hasn't changed, then we don't have to wait. let's exit right away
	if !result.PolicyChanged {
		return
	}

	fmt.Print("Waiting for changes to be applied...")
	var rev *engine.Revision

	var progressBar progress.Indicator
	var progressLast = 0

	finished := retry.Do2(attempts, interval, func() bool {
		var revErr error
		rev, revErr = client.Revision().ShowByPolicy(result.PolicyGeneration)
		if revErr != nil {
			fmt.Print(".")
			return false
		}

		if progressBar == nil {
			fmt.Println()
			progressBar = progress.NewConsole()
			progressBar.SetTotal(rev.Progress.Total)
		}
		for progressLast < rev.Progress.Current {
			progressBar.Advance()
			progressLast++
		}

		return rev.Status != engine.RevisionStatusInProgress
	})

	if !finished {
		progressBar.Done(false)
		fmt.Printf("Timeout! Revision %d has not been applied in %d seconds\n", rev.GetGeneration(), 60*5)
		panic("timeout")
	} else if rev.Status == engine.RevisionStatusSuccess {
		progressBar.Done(true)
		fmt.Printf("Success. Revision %d applied successfully\n", rev.GetGeneration())
	} else if rev.Status == engine.RevisionStatusError {
		progressBar.Done(false)
		fmt.Printf("Error! Revision %d failed with an error and has not been fully applied\n", rev.GetGeneration())
		panic("error")
	}

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
