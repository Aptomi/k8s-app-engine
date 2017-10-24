package client

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/gosuri/uitable"
	"io/ioutil"
	"net/http"
	"time"
)

// Show method retrieves current policy from Aptomi and prints it to console
func Show(cfg *config.Client) error {
	catalog := object.NewCatalog().Append(store.PolicyDataObject)
	cod := yaml.NewCodec(catalog)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, cfg.API.URL()+"/policy", bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	req.Header.Set("Username", cfg.Auth.Username)
	req.Header.Set("Content-Type", "application/yaml")
	req.Header.Set("User-Agent", "aptomictl")

	fmt.Println("Request:", req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint: errcheck

	// todo(slukjanov): process response - check status and print returned data
	fmt.Println("Response:", resp)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from response Body: %s", err))
	}

	objects, err := cod.UnmarshalOneOrMany(respData)
	if err != nil {
		panic(fmt.Sprintf("Error while unmarshalling response: %s", err))
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
