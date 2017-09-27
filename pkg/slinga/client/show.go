package client

import (
	"bytes"
	"fmt"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/gosuri/uitable"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"time"
)

func Show(config *viper.Viper) error {
	catalog := object.NewObjectCatalog(lang.ServiceObject, lang.ContractObject, lang.ClusterObject, lang.RuleObject, lang.DependencyObject)
	cod := yaml.NewCodec(catalog)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	host := config.GetString("server.host")
	port := config.GetInt("server.port")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/api/v1/revision", host, port), bytes.NewBuffer([]byte{}))
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
