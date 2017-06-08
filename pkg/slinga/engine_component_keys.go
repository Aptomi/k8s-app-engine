package slinga

import "strings"

// Create key for the map
func createServiceUsageKey(service *Service, context *Context, allocation *Allocation, component *ServiceComponent) string {
	var componentName string
	if component != nil {
		componentName = component.Name
	} else {
		componentName = componentRootName
	}
	return createServiceUsageKeyFromStr(service.Name, context.Name, allocation.NameResolved, componentName)
}

// Create key for the map
func createServiceUsageKeyFromStr(serviceName string, contextName string, allocationName string, componentName string) string {
	return serviceName + "#" + contextName + "#" + allocationName + "#" + componentName
}

// ParseServiceUsageKey parses key and returns service, component, allocation, component names
func ParseServiceUsageKey(key string) (string, string, string, string) {
	keyArray := strings.Split(key, "#")
	service := keyArray[0]
	context := keyArray[1]
	allocation := keyArray[2]
	component := keyArray[3]
	return service, context, allocation, component
}
