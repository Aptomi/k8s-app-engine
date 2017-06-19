package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

/*
	This file declares all utility structures and methods required for Slinga processing
*/

// Check if context criteria is satisfied
func (context *Context) matches(labels LabelSet) bool {
	return context.Criteria == nil || context.Criteria.allows(labels)
}

// Check if allocation criteria is satisfied
func (allocation *Allocation) matches(labels LabelSet) bool {
	return allocation.Criteria == nil || allocation.Criteria.allows(labels)
}

// Resolve name for an allocation
func (allocation *Allocation) resolveName(user *User, labels LabelSet) error {
	result, err := evaluateTemplate(allocation.Name, user, labels)
	allocation.NameResolved = result
	return err
}

// Whether criteria evaluates to "true" for a given set of labels or not
func (criteria *Criteria) allows(labels LabelSet) bool {
	// If one of the reject criterias matches, then it's not allowed
	for _, reject := range criteria.Reject {
		if evaluate(reject, labels) {
			return false
		}
	}

	// If one of the accept criterias matches, then it's allowed
	for _, reject := range criteria.Accept {
		if evaluate(reject, labels) {
			return true
		}
	}

	// If the accept section is empty, return true
	if len(criteria.Accept) == 0 {
		return true
	}

	return false
}

// Lazily initializes and returns a map of name -> component
func (service *Service) getComponentsMap() map[string]*ServiceComponent {
	if service.componentsMap == nil {
		// Put all components into map
		service.componentsMap = make(map[string]*ServiceComponent)
		for _, c := range service.Components {
			service.componentsMap[c.Name] = c
		}
	}
	return service.componentsMap
}

// Serialize object into YAML
func serializeObject(t interface{}) string {
	d, e := yaml.Marshal(&t)
	if e != nil {
		debug.WithFields(log.Fields{
			"object": t,
			"error":  e,
		}).Fatal("Can't serialize object", e)
	}
	return string(d)
}

// Loads object from YAML file
func loadObjectFromFile(fileName string, data interface{}) interface{} {
	debug.WithFields(log.Fields{
		"file": fileName,
		"type": fmt.Sprintf("%T", data),
	}).Info("Loading entity from file")

	dat, e := ioutil.ReadFile(fileName)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	e = yaml.Unmarshal([]byte(dat), data)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal entity")
	}
	return data
}

// Loads object from YAML file
func loadObjectFromFileDefaultEmpty(fileName string, data interface{}) interface{} {
	debug.WithFields(log.Fields{
		"file": fileName,
		"type": fmt.Sprintf("%T", data),
	}).Info("Loading entity from file")

	dat, e := ioutil.ReadFile(fileName)

	// If the file doesn't exist, it means that DB is empty and we are starting from scratch
	if os.IsNotExist(e) {
		debug.WithFields(log.Fields{
			"file": fileName,
		}).Info("Entity not found. Returning default value")
		return data
	}

	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}

	e = yaml.Unmarshal([]byte(dat), data)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal entity")
	}
	return data
}

// Serialized and stores object in a file
func saveObjectToFile(fileName string, data interface{}) {
	debug.WithFields(log.Fields{
		"file": fileName,
		"type": fmt.Sprintf("%T", data),
	}).Info("Saving entity to file")

	e := ioutil.WriteFile(fileName, []byte(serializeObject(data)), 0644)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to save entity")
	}
}
