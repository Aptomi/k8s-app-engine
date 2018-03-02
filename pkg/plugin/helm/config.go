package helm

import "fmt"

// ClusterConfig represents Kubernetes cluster configuration specific for Helm plugin
type ClusterConfig struct {
	TillerNamespace string `yaml:",omitempty"`
}

func (p *Plugin) parseClusterConfig() error {
	clusterConfig := &ClusterConfig{}
	err := p.cluster.ParseConfigInto(clusterConfig)
	if err != nil {
		return fmt.Errorf("error while parsing helm specific config of cluster %s: %s", p.cluster.Name, err)
	}

	p.tillerNamespace = "kube-system"
	if len(clusterConfig.TillerNamespace) > 0 {
		p.tillerNamespace = clusterConfig.TillerNamespace
	}

	return nil
}
