package plugin

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"sync"
)

const (
	clusterCodeSeparator = "#"
)

type defaultRegistry struct {
	mu sync.Mutex

	clusterTypes       map[string]ClusterPluginConstructor
	codeTypes          map[string]CodePluginConstructor
	postProcessPlugins []PostProcessPlugin

	// Cached plugins instances
	clusterPlugins map[string]ClusterPlugin
	codePlugins    map[string]CodePlugin
}

// NewRegistry creates a registry of aptomi engine plugins
func NewRegistry(clusterTypes map[string]ClusterPluginConstructor, codeTypes map[string]CodePluginConstructor, postProcessPlugins []PostProcessPlugin) Registry {
	return &defaultRegistry{
		clusterTypes:       clusterTypes,
		codeTypes:          codeTypes,
		postProcessPlugins: postProcessPlugins,
		clusterPlugins:     make(map[string]ClusterPlugin),
		codePlugins:        make(map[string]CodePlugin),
	}
}

func (registry *defaultRegistry) ForCluster(cluster *lang.Cluster) (ClusterPlugin, error) {
	constructor, exist := registry.clusterTypes[cluster.Type]
	if !exist {
		return nil, fmt.Errorf("no plugin found for cluster type: %s", cluster.Type)
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	clusterPlugin, exist := registry.clusterPlugins[cluster.Name]
	if !exist {
		clusterPlugin = constructor(cluster)
		registry.clusterPlugins[cluster.Name] = clusterPlugin
	}

	return clusterPlugin, nil
}

func (registry *defaultRegistry) ForCodeType(cluster *lang.Cluster, codeType string) (CodePlugin, error) {
	clusterPlugin, err := registry.ForCluster(cluster)
	if err != nil {
		return nil, err
	}

	constructor, exist := registry.codeTypes[codeType]
	if !exist {
		return nil, fmt.Errorf("no plugin found for code type: %s", codeType)
	}

	key := cluster.Name + clusterCodeSeparator + codeType

	registry.mu.Lock()
	defer registry.mu.Unlock()

	codePlugin, exist := registry.codePlugins[key]
	if !exist {
		codePlugin = constructor(clusterPlugin)
		registry.codePlugins[key] = codePlugin
	}

	return codePlugin, nil
}

func (registry *defaultRegistry) PostProcess() []PostProcessPlugin {
	return registry.postProcessPlugins
}
