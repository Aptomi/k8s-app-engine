package plugin

import "fmt"

type Registry interface {
	GetDeployPlugin(codeType string) (DeployPlugin, error)
	GetClustersPostProcessingPlugins() []ClustersPostProcessPlugin
}

type defaultRegistry struct {
	deployPlugins              map[string]DeployPlugin
	clustersPostProcessPlugins []ClustersPostProcessPlugin
}

func NewRegistry(deployPlugins []DeployPlugin, clustersPostProcessPlugins []ClustersPostProcessPlugin) Registry {
	deployPluginsMap := make(map[string]DeployPlugin, len(deployPlugins))
	for _, plugin := range deployPlugins {
		for _, codeType := range plugin.GetSupportedCodeTypes() {
			if _, exist := deployPluginsMap[codeType]; exist {
				// todo(slukjanov): is it ok to panic here?
				panic("More than one plugin registered for the same codeType: " + codeType)
			}
			deployPluginsMap[codeType] = plugin
		}
	}
	return &defaultRegistry{
		deployPlugins:              deployPluginsMap,
		clustersPostProcessPlugins: clustersPostProcessPlugins,
	}
}

func (reg *defaultRegistry) GetDeployPlugin(codeType string) (DeployPlugin, error) {
	plugin, exist := reg.deployPlugins[codeType]
	if !exist {
		return nil, fmt.Errorf("Can't find deploy plugin for codeType: %s", codeType)
	}
	return plugin, nil
}

func (reg *defaultRegistry) GetClustersPostProcessingPlugins() []ClustersPostProcessPlugin {
	return reg.clustersPostProcessPlugins
}
