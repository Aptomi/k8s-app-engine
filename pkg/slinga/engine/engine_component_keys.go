package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"strings"
)

// componentInstanceKeySeparator is a separator between strings in ComponentInstanceKey
const componentInstanceKeySeparator = "#"

// componentUnresolvedName is placeholder for unresolved entries
const componentUnresolvedName = "unknown"

// componentRootName is a name of component for service entry (which in turn consists of components)
const componentRootName = "root"

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

// IsService returns 'true' if it's a service instance key and we can't go up anymore. And it will return 'false' if it's a component instance key
func (cik *ComponentInstanceKey) IsService() bool {
	return cik.ComponentName == componentRootName
}

// IsComponent returns 'true' if it's a component instance key and we can go up to the corresponding service. And it will return 'false' if it's a service instance key
func (cik *ComponentInstanceKey) IsComponent() bool {
	return cik.ComponentName != componentRootName
}

// GetParentServiceKey returns a key for the parent service, replacing componentName with componentRootName
func (cik *ComponentInstanceKey) GetParentServiceKey() *ComponentInstanceKey {
	if cik.ComponentName == componentRootName {
		return cik
	}
	serviceCik := cik.MakeCopy()
	serviceCik.ComponentName = componentRootName
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
			}, componentInstanceKeySeparator)
	}
	return cik.key
}

// Returns allocation name combined with allocation keys
func (cik ComponentInstanceKey) GetAllocationNameWithKeys() string {
	result := cik.AllocationName
	if len(cik.AllocationsKeysResolved) > 0 {
		result += componentInstanceKeySeparator + strings.Join(cik.AllocationsKeysResolved, componentInstanceKeySeparator)
	}
	return result
}

// If context has not been resolved and we need a key, generate one
func getContextNameUnsafe(context *Context) string {
	if context == nil {
		return componentUnresolvedName
	}
	return context.GetName()
}

// If allocation has not been resolved and we need a key, generate one
func getAllocationNameUnsafe(context *Context) string {
	if context == nil || context.Allocation == nil || len(context.Allocation.Name) <= 0 {
		return componentUnresolvedName
	}
	return context.Allocation.Name
}

// If component has not been resolved and we need a key, generate one
func getComponentNameUnsafe(component *ServiceComponent) string {
	if component == nil {
		return componentRootName
	}
	return component.Name
}
