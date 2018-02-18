package helm

import "fmt"

// ClusterConfig represents Kubernetes cluster configuration specific for Helm plugin
type ClusterConfig struct {
	TillerNamespace string `yaml:",omitempty"`
}

func (plugin *Plugin) parseClusterConfig() error {
	clusterConfig := &ClusterConfig{}
	err := plugin.cluster.ParseConfigInto(clusterConfig)
	if err != nil {
		err = fmt.Errorf("error while parsing helm specific config of cluster %s: %s", plugin.cluster.Name, err)
	}

	plugin.tillerNamespace = "kube-system"
	if len(clusterConfig.TillerNamespace) > 0 {
		plugin.tillerNamespace = clusterConfig.TillerNamespace
	}

	return nil
}
