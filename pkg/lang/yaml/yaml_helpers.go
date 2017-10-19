package yaml

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// DeserializeObject deserializes YAML into object
func DeserializeObject(s string, object interface{}) error {
	return yaml.Unmarshal([]byte(s), object)
}

// SerializeObject serializes object into YAML
func SerializeObject(t interface{}) string {
	d, err := yaml.Marshal(&t)
	if err != nil {
		panic(fmt.Sprintf("Can't serialize object '%+v': %s", t, err))
	}
	return string(d)
}

// LoadObjectFromFile loads object from YAML file
func LoadObjectFromFile(fileName string, data interface{}) interface{} {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("Unable to read file '%s': %s", fileName, err))
	}
	err = yaml.Unmarshal(dat, data)
	if err != nil {
		panic(fmt.Sprintf("Unable to unmarshal entity from '%s': %s", fileName, err))
	}
	return data
}

// LoadObjectFromFileDefaultEmpty loads object from YAML file
func LoadObjectFromFileDefaultEmpty(fileName string, data interface{}) interface{} {
	dat, err := ioutil.ReadFile(fileName)

	// If the file doesn't exist, it means that DB is empty and we are starting from scratch
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

// SaveObjectToFile serializes and stores object in a file
func SaveObjectToFile(fileName string, data interface{}) {
	err := ioutil.WriteFile(fileName, []byte(SerializeObject(data)), 0644)
	if err != nil {
		panic(fmt.Sprintf("Unable to save entity to '%s': %s", fileName, err))
	}
}
