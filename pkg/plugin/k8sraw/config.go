package k8sraw

import "fmt"

// ClusterConfig represents Kubernetes cluster configuration specific for K8s raw plugin
type ClusterConfig struct {
	DataNamespace string `yaml:",omitempty"`
}

func (plugin *Plugin) parseClusterConfig() error {
	clusterConfig := &ClusterConfig{}
	err := plugin.cluster.ParseConfigInto(clusterConfig)
	if err != nil {
		return fmt.Errorf("error while parsing k8s raw specific config of cluster %s: %s", plugin.cluster.Name, err)
	}

	plugin.dataNamespace = "aptomi"
	if len(plugin.config.DataNamespace) > 0 {
		plugin.dataNamespace = clusterConfig.DataNamespace
	}
	if len(clusterConfig.DataNamespace) > 0 {
		plugin.dataNamespace = clusterConfig.DataNamespace
	}

	return nil
}
