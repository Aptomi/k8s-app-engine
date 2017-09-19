package resolve

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
// When adding keys to this method, don't forget to modify the constructor and copy routines
type ComponentInstanceKey struct {
	// cached version of component key
	key string

	// required fields
	ClusterName         string // mandatory
	ContractName        string // mandatory
	ContextName         string // mandatory
	ContextNameWithKeys string // calculated
	ServiceName         string // determined from the context (included into key for readability)
	ComponentName       string // component name
}

// NewComponentInstanceKey creates a new ComponentInstanceKey
func NewComponentInstanceKey(cluster *Cluster, contract *Contract, context *Context, allocationsKeysResolved []string, service *Service, component *ServiceComponent) *ComponentInstanceKey {
	contextName := getContextNameUnsafe(context)
	contextNameWithKeys := getContextNameWithKeys(contextName, allocationsKeysResolved)
	return &ComponentInstanceKey{
		ClusterName:         getClusterNameUnsafe(cluster),
		ContractName:        getContractNameUnsafe(contract),
		ContextName:         contextName,
		ContextNameWithKeys: contextNameWithKeys,
		ServiceName:         getServiceNameUnsafe(service),
		ComponentName:       getComponentNameUnsafe(component),
	}
}

// MakeCopy creates a copy of ComponentInstanceKey
func (cik *ComponentInstanceKey) MakeCopy() *ComponentInstanceKey {
	return &ComponentInstanceKey{
		ClusterName:         cik.ClusterName,
		ContractName:        cik.ContractName,
		ContextName:         cik.ContextName,
		ContextNameWithKeys: cik.ContextNameWithKeys,
		ComponentName:       cik.ComponentName,
	}
}

// IsService returns 'true' if it's a contract instance key and we can't go up anymore. And it will return 'false' if it's a component instance key
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
				cik.ClusterName,
				cik.ContractName,
				cik.ContextNameWithKeys,
				cik.ComponentName,
			}, componentInstanceKeySeparator)
	}
	return cik.key
}

// If cluster has not been resolved yet and we need a key, generate one
// Otherwise use cluster name
func getClusterNameUnsafe(cluster *Cluster) string {
	if cluster == nil {
		return componentUnresolvedName
	}
	return cluster.Name
}

// If contract has not been resolved yet and we need a key, generate one
// Otherwise use contract name
func getContractNameUnsafe(contract *Contract) string {
	if contract == nil {
		return componentUnresolvedName
	}
	return contract.Name
}

// If context has not been resolved yet and we need a key, generate one
// Otherwise use context name
func getContextNameUnsafe(context *Context) string {
	if context == nil {
		return componentUnresolvedName
	}
	return context.Name
}

// If service has not been resolved yet and we need a key, generate one
// Otherwise use service name
func getServiceNameUnsafe(service *Service) string {
	if service == nil {
		return componentUnresolvedName
	}
	return service.Name
}

// If component has not been resolved yet and we need a key, generate one
// Otherwise use component name
func getComponentNameUnsafe(component *ServiceComponent) string {
	if component == nil {
		return componentRootName
	}
	return component.Name
}

// Returns context name combined with allocation keys
func getContextNameWithKeys(contextName string, allocationKeysResolved []string) string {
	result := contextName
	if len(allocationKeysResolved) > 0 {
		result += componentInstanceKeySeparator + strings.Join(allocationKeysResolved, componentInstanceKeySeparator)
	}
	return result
}
