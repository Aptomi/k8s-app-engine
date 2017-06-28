package yaml

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	. "github.com/Frostman/aptomi/pkg/slinga/log"
)

// SerializeObject serializes object into YAML
func SerializeObject(t interface{}) string {
	d, e := yaml.Marshal(&t)
	if e != nil {
		Debug.WithFields(log.Fields{
			"object": t,
			"error":  e,
		}).Panic("Can't serialize object", e)
	}
	return string(d)
}

// Loads object from YAML file
func LoadObjectFromFile(fileName string, data interface{}) interface{} {
	Debug.WithFields(log.Fields{
		"file": fileName,
		"type": fmt.Sprintf("%T", data),
	}).Info("Loading entity from file")

	dat, e := ioutil.ReadFile(fileName)
	if e != nil {
		Debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Panic("Unable to read file")
	}
	e = yaml.Unmarshal([]byte(dat), data)
	if e != nil {
		Debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Panic("Unable to unmarshal entity")
	}
	return data
}

// LoadObjectFromFileDefaultEmpty loads object from YAML file
func LoadObjectFromFileDefaultEmpty(fileName string, data interface{}) interface{} {
	Debug.WithFields(log.Fields{
		"file": fileName,
		"type": fmt.Sprintf("%T", data),
	}).Info("Loading entity from file")

	dat, e := ioutil.ReadFile(fileName)

	// If the file doesn't exist, it means that DB is empty and we are starting from scratch
	if os.IsNotExist(e) {
		Debug.WithFields(log.Fields{
			"file": fileName,
		}).Info("Entity not found. Returning default value")
		return data
	}

	if e != nil {
		Debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Panic("Unable to read file")
	}

	e = yaml.Unmarshal([]byte(dat), data)
	if e != nil {
		Debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Panic("Unable to unmarshal entity")
	}
	return data
}

// SaveObjectToFile serializes and stores object in a file
func SaveObjectToFile(fileName string, data interface{}) {
	Debug.WithFields(log.Fields{
		"file": fileName,
		"type": fmt.Sprintf("%T", data),
	}).Info("Saving entity to file")

	e := ioutil.WriteFile(fileName, []byte(SerializeObject(data)), 0644)
	if e != nil {
		Debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Panic("Unable to save entity")
	}
}
