package yaml

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// SerializeObject serializes object into YAML
func SerializeObject(t interface{}) string {
	d, err := yaml.Marshal(&t)
	if err != nil {
		panic(fmt.Sprintf("Can't serialize object '%+v': %s", t, err))
	}
	return string(d)
}

// LoadObjectFromFileDefaultEmpty loads object from YAML file
func LoadObjectFromFileDefaultEmpty(fileName string, data interface{}) interface{} {
	dat, err := ioutil.ReadFile(fileName)

	// If file doesn't exist, return empty data
	if os.IsNotExist(err) {
		return data
	}

	if err != nil {
		panic(fmt.Sprintf("Unable to read file '%s': %s", fileName, err))
	}

	err = yaml.Unmarshal(dat, data)
	if err != nil {
		panic(fmt.Sprintf("Unable to unmarshal entity from '%s': %s", fileName, err))
	}
	return data
}
