package slinga

import (
	"strings"
	. "github.com/Frostman/aptomi/pkg/slinga/language"
)

// ComponentUnresolvedName is placeholder for unresolved entries
const ComponentUnresolvedName = "unknown"

// ComponentRootName is a name of component for service entry (which in turn consists of components)
const ComponentRootName = "root"

// If context has not been resolved and we need a key, generate one
func getContextNameUnsafe(context *Context) string {
	if context == nil {
		return ComponentUnresolvedName
	}
	return context.Name
}

// If allocation has not been resolved and we need a key, generate one
func getAllocationNameUnsafe(allocation *Allocation) string {
	if allocation == nil || len(allocation.NameResolved) <= 0 {
		return ComponentUnresolvedName
	}
	return allocation.NameResolved
}

// If component has not been resolved and we need a key, generate one
func getComponentNameUnsafe(component *ServiceComponent) string {
	if component == nil {
		return ComponentRootName
	}
	return component.Name
}

// Create key for the map
func createServiceUsageKey(serviceName string, context *Context, allocation *Allocation, component *ServiceComponent) string {
	return createServiceUsageKeyFromStr(serviceName, getContextNameUnsafe(context), getAllocationNameUnsafe(allocation), getComponentNameUnsafe(component))
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
