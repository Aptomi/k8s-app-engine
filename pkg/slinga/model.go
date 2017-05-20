package slinga

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
)

/*
	This file declares all utility structures and methods required for Slinga processing
*/

// Set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// Apply set of transformations to labels
func (user *User) getLabelSet() LabelSet {
	return LabelSet{Labels: user.Labels}
}

// Apply set of transformations to labels
func (src *LabelSet) applyTransform(ops *LabelOperations) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v
	}

	// set labels
	for k, v := range (*ops)["set"] {
		result.Labels[k] = v
	}

	// remove labels
	for k, _ := range (*ops)["remove"] {
		delete(result.Labels, k)
	}

	return result
}

// Merge two sets of labels
func (src LabelSet) addLabels(ops LabelSet) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v
	}

	// put new labels
	for k, v := range ops.Labels {
		result.Labels[k] = v
	}

	return result
}

// Check if context criteria is satisfied
func (context *Context) matches(labels LabelSet) bool {
	for _, c := range context.Criteria {
		if evaluate(c, labels) {
			return true
		}
	}
	return false
}

// Check if allocation criteria is satisfied
func (allocation *Allocation) matches(labels LabelSet) bool {
	for _, c := range allocation.Criteria {
		if evaluate(c, labels) {
			return true
		}
	}
	return false
}

// Resolve name for an allocation
func (allocation *Allocation) resolveName(user User) error {
	result, err := evaluateTemplate(allocation.Name, user)
	allocation.NameResolved = result
	return err
}

// Evaluates a template
func evaluateTemplate(templateStr string, user User) (string, error) {

	type Parameters struct {
		User User
	}
	param := Parameters{User: user}

	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", errors.New("Invalid template " + templateStr)
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, param)

	if err != nil {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	return doc.String(), nil
}

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
