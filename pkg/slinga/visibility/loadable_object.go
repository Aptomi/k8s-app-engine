package visibility

import (
	"reflect"
	"github.com/Frostman/aptomi/pkg/slinga"
	"strings"
)

type loadableObject interface {
	isItMyID(string) string
	getDetails(string, slinga.ServiceUsageState) interface{}
}

func getLoadableObject(id string) loadableObject {
	var registeredObjects = []reflect.Type {
		reflect.TypeOf(dependencyNode{}),
		reflect.TypeOf(serviceNode{}),
		reflect.TypeOf(serviceInstanceNode{}),
	}

	for _, t := range registeredObjects {
		v := reflect.New(t).Interface().(loadableObject)
		if len(v.isItMyID(id)) > 0 {
			return v
		}
	}
	return nil
}

func cutPrefixOrEmpty(s string, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return strings.TrimPrefix(s, prefix)
	}
	return ""
}
