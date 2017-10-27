package client

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/actioninfo"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/codec/yaml"
	"github.com/gosuri/uitable"
	"io/ioutil"
	"net/http"
	"time"
)

// Endpoints retrieves all endpoints for deployed services
func Endpoints(cfg *config.Client) error {
	catalog := object.NewCatalog().Append(actioninfo.Objects...)
	cod := yaml.NewCodec(catalog)

	client := &http.Client{
		// todo make configurable
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, cfg.API.URL()+"/endpoints", nil)
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

	// todo bad logging
	fmt.Println("Response data:\n" + string(respData))

	objects, err := cod.UnmarshalOneOrMany(respData)
	if err != nil {
		panic(fmt.Sprintf("Error while unmarshalling response: %s", err))
	}

	// todo(slukjanov): pretty print response
	table := uitable.New()
	//table.MaxColWidth = 50

	table.AddRow("#", "Namespace", "Kind", "Name", "Generation", "Object")
	for idx, obj := range objects {
		table.AddRow(idx, obj.GetNamespace(), obj.GetKind(), obj.GetName(), obj.GetGeneration(), obj)
	}
	fmt.Println(table)

	return nil
}
