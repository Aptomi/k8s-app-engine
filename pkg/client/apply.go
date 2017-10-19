package client

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/codec"
	"github.com/Aptomi/aptomi/pkg/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/gosuri/uitable"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// todo make a client object with apply/show methods

// Apply finds all policy files and uploads/applies them by making a call to aptomi server
func Apply(cfg *config.Client) error {
	catalog := object.NewCatalog().Append(lang.Objects...).Append(store.Objects...).Append(engine.ActionObjects...)
	cod := yaml.NewCodec(catalog)

	allObjects, err := readFiles(cfg.Apply.PolicyPaths, cod)
	if err != nil {
		return err
	}

	// todo(slukjanov): here we can use some more efficient marshaller like gob, etc
	// prepare single []byte to send to API
	data, err := cod.MarshalMany(allObjects)
	if err != nil {
		return fmt.Errorf("error while marshaling data for sending to API: %s", err)
	}

	client := &http.Client{
		// todo make configurable
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, cfg.API.URL()+"/policy", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/yaml")
	req.Header.Set("User-Agent", "aptomictl")

	fmt.Println("Request:", req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint: errcheck

	// todo(slujanov): process response - check status and print returned data
	fmt.Println("Response:", resp)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from response Body: %s", err))
	}

	// todo bad logging
	fmt.Println("Response data: " + string(respData))

	objects, err := cod.UnmarshalOneOrMany(respData)
	if err != nil {
		panic(fmt.Sprintf("Error while unmarshaling response: %s", err))
	}

	// todo(slukjanov): pretty print response
	table := uitable.New()
	table.MaxColWidth = 50

	table.AddRow("#", "Namespace", "Kind", "Name", "Generation", "Object")
	for idx, obj := range objects {
		table.AddRow(idx, obj.GetNamespace(), obj.GetKind(), obj.GetName(), obj.GetGeneration(), obj)
	}
	fmt.Println(table)

	return nil
}

func readFiles(policyPaths []string, codec codec.MarshallerUnmarshaller) ([]object.Base, error) {
	files, err := findPolicyFiles(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("error while searching for policy files: %s", err)
	}

	allObjects := make([]object.Base, 0)
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("can't read file %s error: %s", file, err)
		}

		// todo(slukjanov): here we can try multiple marshalers, toml for example
		objects, err := codec.UnmarshalOneOrMany(data)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal file %s error: %s", file, err)
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

	sort.Strings(allFiles) // todo(slukjanov): do we really need to sort files?

	// todo(slukjanov): log list of files from which we're applying policy
	//fmt.Println("Apply policy from following files:")
	//for idx, policyPath := range allFiles {
	//	fmt.Println(idx, "-", policyPath)
	//}

	return allFiles, nil
}
