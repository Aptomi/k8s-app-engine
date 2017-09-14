package client

import (
	"bytes"
	"fmt"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/gosuri/uitable"
	"github.com/mattn/go-zglob"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func Apply(config *viper.Viper) error {
	policyPaths := config.GetStringSlice("apply.policyPaths")

	catalog := object.NewObjectCatalog(lang.ServiceObject, lang.ContractObject, lang.ContextObject, lang.ClusterObject, lang.RuleObject, lang.DependencyObject)
	cod := yaml.NewCodec(catalog)

	allObjects, err := readFiles(policyPaths, cod)
	if err != nil {
		return err
	}

	// todo(slukjanov): here we can use some more efficient marshaller like gob, etc
	// prepare single []byte to send to API
	data, err := cod.MarshalMany(allObjects)
	if err != nil {
		return fmt.Errorf("Error while marshaling data for sending to API: %s", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	host := config.GetString("server.host")
	port := config.GetInt("server.port")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/api/v1/revision", host, port), bytes.NewBuffer(data))
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
	defer resp.Body.Close()

	// todo(slujanov): process response - check status and print returned data
	fmt.Println("Response:", resp)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from response Body: %s", err))
	}

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

func readFiles(policyPaths []string, codec codec.MarshalUnmarshaler) ([]object.Base, error) {
	files, err := findPolicyFiles(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("Error while searching for policy files: %s", err)
	}

	allObjects := make([]object.Base, 0)
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("Can't read file %s error: %s", file, err)
		}

		// todo(slukjanov): here we can try multiple marshalers, toml for example
		objects, err := codec.UnmarshalOneOrMany(data)
		if err != nil {
			return nil, fmt.Errorf("Can't unmarshal file %s error: %s", file, err)
		}

		allObjects = append(allObjects, objects...)
	}

	if len(allObjects) == 0 {
		return nil, fmt.Errorf("No objects found in %s", policyPaths)
	}

	return allObjects, nil
}

func findPolicyFiles(policyPaths []string) ([]string, error) {
	allFiles := make([]string, 0, len(policyPaths))

	for _, rawPolicyPath := range policyPaths {
		policyPath, err := filepath.Abs(rawPolicyPath)
		if err != nil {
			return nil, fmt.Errorf("Error reading filepath: %s", err)
		}

		if stat, err := os.Stat(policyPath); err == nil {
			if stat.IsDir() { // if dir provided, use all yaml files from it
				files, err := zglob.Glob(filepath.Join(policyPath, "**", "*.yaml"))
				if err != nil {
					return nil, fmt.Errorf("Error while searching yaml files in directory: %s error: %s", policyPath, err)
				}
				allFiles = append(allFiles, files...)
			} else { // if specific file provided, use it
				allFiles = append(allFiles, policyPath)
			}
		} else if os.IsNotExist(err) {
			return nil, fmt.Errorf("Path doesn't exists: %s error: %s", policyPath, err)
		} else {
			return nil, fmt.Errorf("Error while processing path: %s", err)
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

/*
c := &http.Client{
Timeout: 15 * time.Second,
}
resp, err := c.Get("https://blog.filippo.io/")
*/
