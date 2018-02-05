package plugin

import "fmt"

// Registry is a registry of all Aptomi engine plugins
type Registry interface {
	GetDeployPlugin(codeType string) (DeployPlugin, error)
	GetPostProcessingPlugins() []PostProcessPlugin
}

// RegistryFactory returns plugins registry on demand
type RegistryFactory func() Registry

type defaultRegistry struct {
	deployPlugins      map[string]DeployPlugin
	postProcessPlugins []PostProcessPlugin
}

// NewRegistry creates a registry of aptomi engine plugins
func NewRegistry(deployPlugins []DeployPlugin, postProcessPlugins []PostProcessPlugin) Registry {
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
		deployPlugins:      deployPluginsMap,
		postProcessPlugins: postProcessPlugins,
	}
}

func (reg *defaultRegistry) GetDeployPlugin(codeType string) (DeployPlugin, error) {
	plugin, exist := reg.deployPlugins[codeType]
	if !exist {
		return nil, fmt.Errorf("can't find deploy plugin for codeType: %s", codeType)
	}
	return plugin, nil
}

func (reg *defaultRegistry) GetPostProcessingPlugins() []PostProcessPlugin {
	return reg.postProcessPlugins
}
