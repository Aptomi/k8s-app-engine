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
	"github.com/mattn/go-zglob"
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

	log.Info("Applying policy from:")
	for _, policyPath := range allFiles {
		log.Infof("  [*] %s", policyPath)
	}

	return allFiles, nil
}

func waitForApplyToFinish(attempts int, interval time.Duration, client client.Core, result *api.PolicyUpdateResult) {
	fmt.Print("Waiting for updated policy to be applied...")
	time.Sleep(interval)

	var progressBar progress.Indicator
	var progressLast = 0

	var rev *engine.Revision
	finished := retry.Do(attempts, interval, func() bool {
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
		fmt.Printf("Timeout. Revision %d has not been applied in %d seconds\n", rev.GetGeneration(), 60*5)
		panic("timeout")
	} else if rev.Status == engine.RevisionStatusSuccess {
		progressBar.Done(true)
		fmt.Printf("Success! Revision %d created and applied\n", rev.GetGeneration())
	} else if rev.Status == engine.RevisionStatusError {
		progressBar.Done(false)
		fmt.Printf("Error. Revision %d failed with an error and has not been fully applied\n", rev.GetGeneration())
		panic("error")
	}

}
