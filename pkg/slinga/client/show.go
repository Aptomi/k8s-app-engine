package client

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/config"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/gosuri/uitable"
	"io/ioutil"
	"net/http"
	"time"
)

// Show method retrieves current policy from aptomi and prints it
func Show(cfg *config.Client) error {
	catalog := object.NewCatalog().Append(lang.Objects...)
	cod := yaml.NewCodec(catalog)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, cfg.Server.URL(), bytes.NewBuffer([]byte{}))
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
