package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"strings"
)

// ComponentInstanceKeySeparator is a separator between strings in ComponentInstanceKey
const ComponentInstanceKeySeparator = "#"

// ComponentUnresolvedName is placeholder for unresolved entries
const ComponentUnresolvedName = "unknown"

// ComponentRootName is a name of component for service entry (which in turn consists of components)
const ComponentRootName = "root"

// ComponentInstanceKey is a struct representing a key for the component instance and the fields it consists of
type ComponentInstanceKey struct {
	// cached version of component key
	key string

	// required fields
	ServiceName    string
	ContextName    string
	AllocationName string
	ComponentName  string

	// additional allocation keys
	AllocationsKeysResolved []string
}

// NewComponentInstanceKey creates a new ComponentInstanceKey
func NewComponentInstanceKey(serviceName string, context *Context, allocationsKeysResolved []string, component *ServiceComponent) *ComponentInstanceKey {
	return &ComponentInstanceKey{
		ServiceName:             serviceName,
		ContextName:             getContextNameUnsafe(context),
		AllocationName:          getAllocationNameUnsafe(context),
		AllocationsKeysResolved: allocationsKeysResolved,
		ComponentName:           getComponentNameUnsafe(component),
	}
}

// MakeCopy creates a copy of ComponentInstanceKey
func (cik *ComponentInstanceKey) MakeCopy() *ComponentInstanceKey {
	return &ComponentInstanceKey{
		ServiceName:             cik.ServiceName,
		ContextName:             cik.ContextName,
		AllocationName:          cik.AllocationName,
		ComponentName:           cik.ComponentName,
		AllocationsKeysResolved: cik.AllocationsKeysResolved,
	}
}

// GetParentServiceKey returns a key for the parent service, replacing componentName with ComponentRootName
func (cik *ComponentInstanceKey) GetParentServiceKey() *ComponentInstanceKey {
	if cik.ComponentName == ComponentRootName {
		return cik
	}
	serviceCik := cik.MakeCopy()
	serviceCik.ComponentName = ComponentRootName
	return serviceCik
}

// GetKey returns a string key
func (cik ComponentInstanceKey) GetKey() string {
	if cik.key == "" {
		cik.key = strings.Join(
			[]string{
				cik.ServiceName,
				cik.ContextName,
				cik.GetAllocationNameWithKeys(),
				cik.ComponentName,
			}, ComponentInstanceKeySeparator)
	}
	return cik.key
}

// Returns allocation name combined with allocation keys
func (cik ComponentInstanceKey) GetAllocationNameWithKeys() string {
	result := cik.AllocationName
	if len(cik.AllocationsKeysResolved) > 0 {
		result += ComponentInstanceKeySeparator + strings.Join(cik.AllocationsKeysResolved, ComponentInstanceKeySeparator)
	}
	return result
}

// If context has not been resolved and we need a key, generate one
func getContextNameUnsafe(context *Context) string {
	if context == nil {
		return ComponentUnresolvedName
	}
	return context.GetName()
}

// If allocation has not been resolved and we need a key, generate one
func getAllocationNameUnsafe(context *Context) string {
	if context == nil || context.Allocation == nil || len(context.Allocation.Name) <= 0 {
		return ComponentUnresolvedName
	}
	return context.Allocation.Name
}

// If component has not been resolved and we need a key, generate one
func getComponentNameUnsafe(component *ServiceComponent) string {
	if component == nil {
		return ComponentRootName
	}
	return component.Name
}
