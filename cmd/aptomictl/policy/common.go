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

	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("error while getting info about stdin")
	}

	// if used as something | aptomictl
	if info.Mode()&os.ModeNamedPipe != 0 {
		return readLangObjectsFromStdin(codec)
	}

	return readLangObjectsFromFiles(policyPaths, codec)
}

func readLangObjectsFromStdin(codec runtime.Codec) ([]runtime.Object, error) {
	log.Info("Applying policy from stdin (or pipe)")
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
	if len(policyPaths) <= 0 {
		return nil, fmt.Errorf("policy file path is not specified")
	}

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

		return nil, fmt.Errorf("path doesn't exist or no YAML files found under: %s", policyPath)
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
