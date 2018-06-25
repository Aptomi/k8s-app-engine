package resolve

import (
	"encoding/base32"
	"hash/fnv"
	"strings"

	"github.com/Aptomi/aptomi/pkg/lang"
)

// componentInstanceKeySeparator is a separator between strings in ComponentInstanceKey
const componentInstanceKeySeparator = "#"

// componentUnresolvedName is placeholder for unresolved entries
const componentUnresolvedName = "unknown"

// componentRootName is a name of component for bundle entry (which in turn consists of components)
const componentRootName = "root"

// ComponentInstanceKey is a key for component instance. During policy resolution every component instance gets
// assigned a unique string key. It's important to form those keys correctly, so that we can make actual comparison
// of actual state (components with their keys) and desired state (components with their keys).
//
// Currently, component keys are formed from multiple parameters as follows.
// Cluster gets included as a part of the key (components running on different clusters must have different keys).
// Namespace gets included as a part of the key (components from different namespaces must have different keys).
// Service, Context (with allocation keys), Bundle get included as a part of the key (Bundle must be within the same namespace as Service).
// ComponentName gets included as a part of the key. For bundle-level component instances, ComponentName is
// set to componentRootName, while for all component instances within a bundle an actual Component.Name is used.
type ComponentInstanceKey struct {
	// cached version of component key
	key string

	// required fields
	ClusterNameSpace    string // mandatory
	ClusterName         string // mandatory
	TargetSuffix        string // mandatory
	Namespace           string // determined from the service
	ServiceName         string // mandatory
	ContextName         string // mandatory
	KeysResolved        string // mandatory
	ContextNameWithKeys string // calculated
	BundleName          string // determined from the context (included into key for readability)
	ComponentName       string // component name
}

// NewComponentInstanceKey creates a new ComponentInstanceKey
func NewComponentInstanceKey(cluster *lang.Cluster, targetSuffix string, service *lang.Service, context *lang.Context, allocationKeysResolved []string, bundle *lang.Bundle, component *lang.BundleComponent) *ComponentInstanceKey {
	contextName := getContextNameUnsafe(context)
	keysResolved := strings.Join(allocationKeysResolved, componentInstanceKeySeparator)
	contextNameWithKeys := contextName
	if len(keysResolved) > 0 {
		contextNameWithKeys = strings.Join([]string{contextNameWithKeys, keysResolved}, componentInstanceKeySeparator)
	}
	if len(targetSuffix) <= 0 {
		targetSuffix = componentUnresolvedName
	}
	return &ComponentInstanceKey{
		ClusterNameSpace:    getClusterNamespaceUnsafe(cluster),
		ClusterName:         getClusterNameUnsafe(cluster),
		TargetSuffix:        targetSuffix,
		Namespace:           getServiceNamespaceUnsafe(service),
		ServiceName:         getServiceNameUnsafe(service),
		ContextName:         contextName,
		KeysResolved:        keysResolved,
		ContextNameWithKeys: contextNameWithKeys,
		BundleName:          getBundleNameUnsafe(bundle),
		ComponentName:       getComponentNameUnsafe(component),
	}
}

// MakeCopy creates a copy of ComponentInstanceKey
func (cik *ComponentInstanceKey) MakeCopy() *ComponentInstanceKey {
	return &ComponentInstanceKey{
		ClusterNameSpace:    cik.ClusterNameSpace,
		ClusterName:         cik.ClusterName,
		TargetSuffix:        cik.TargetSuffix,
		Namespace:           cik.Namespace,
		ServiceName:         cik.ServiceName,
		ContextName:         cik.ContextName,
		KeysResolved:        cik.KeysResolved,
		ContextNameWithKeys: cik.ContextNameWithKeys,
		ComponentName:       cik.ComponentName,
	}
}

// IsBundle returns 'true' if it's a bundle instance key and we can't go up anymore. And it will return 'false' if it's a component instance key
func (cik *ComponentInstanceKey) IsBundle() bool {
	return cik.ComponentName == componentRootName
}

// IsComponent returns 'true' if it's a component instance key and we can go up to the corresponding bundle. And it will return 'false' if it's a bundle instance key
func (cik *ComponentInstanceKey) IsComponent() bool {
	return cik.ComponentName != componentRootName
}

// GetParentBundleKey returns a key for the parent bundle, replacing componentName with componentRootName
func (cik *ComponentInstanceKey) GetParentBundleKey() *ComponentInstanceKey {
	if cik.ComponentName == componentRootName {
		return cik
	}
	bundleCik := cik.MakeCopy()
	bundleCik.ComponentName = componentRootName
	return bundleCik
}

// GetKey returns a string key
func (cik ComponentInstanceKey) GetKey() string {
	if cik.key == "" {
		cik.key = strings.Join(
			[]string{
				cik.ClusterNameSpace,
				cik.ClusterName,
				cik.TargetSuffix,
				cik.Namespace,
				cik.ServiceName,
				cik.ContextNameWithKeys,
				cik.ComponentName,
			}, componentInstanceKeySeparator)
	}
	return cik.key
}

var (
	base32LowerCaseHexEncoding = base32.NewEncoding("0123456789abcdefghijklmnopqrstuv")
)

// GetDeployName returns a string that could be used as name for deployment inside the cluster
func (cik ComponentInstanceKey) GetDeployName() string {
	h := fnv.New64a()
	_, err := h.Write([]byte(cik.GetKey()))
	if err != nil {
		panic(err)
	}
	keyHash := base32LowerCaseHexEncoding.EncodeToString(h.Sum(nil))[0:13]

	return "a-" + keyHash
}

// If cluster has not been resolved yet and we need a key, generate one
// Otherwise use cluster name
func getClusterNameUnsafe(cluster *lang.Cluster) string {
	if cluster == nil {
		return componentUnresolvedName
	}
	return cluster.Name
}

// If cluster has not been resolved yet and we need a key, generate one
// Otherwise use cluster space
func getClusterNamespaceUnsafe(cluster *lang.Cluster) string {
	if cluster == nil {
		return componentUnresolvedName
	}
	return cluster.Namespace
}

// If service has not been resolved yet and we need a key, generate one
// Otherwise use service name
func getServiceNameUnsafe(service *lang.Service) string {
	if service == nil {
		return componentUnresolvedName
	}
	return service.Name
}

// If service has not been resolved yet and we need a key, generate one
// Otherwise use service namespace
func getServiceNamespaceUnsafe(service *lang.Service) string {
	if service == nil {
		return componentUnresolvedName
	}
	return service.Namespace
}

// If context has not been resolved yet and we need a key, generate one
// Otherwise use context name
func getContextNameUnsafe(context *lang.Context) string {
	if context == nil {
		return componentUnresolvedName
	}
	return context.Name
}

// If bundle has not been resolved yet and we need a key, generate one
// Otherwise use bundle name
func getBundleNameUnsafe(bundle *lang.Bundle) string {
	if bundle == nil {
		return componentUnresolvedName
	}
	return bundle.Name
}

// If component has not been resolved yet and we need a key, generate one
// Otherwise use component name
func getComponentNameUnsafe(component *lang.BundleComponent) string {
	if component == nil {
		return componentRootName
	}
	return component.Name
}
