package k8sraw

import "fmt"

// ClusterConfig represents Kubernetes cluster configuration specific for K8s raw plugin
type ClusterConfig struct {
	DataNamespace string `yaml:",omitempty"`
}

func (p *Plugin) parseClusterConfig() error {
	clusterConfig := &ClusterConfig{}
	err := p.cluster.ParseConfigInto(clusterConfig)
	if err != nil {
		return fmt.Errorf("error while parsing k8s raw specific config of cluster %s: %s", p.cluster.Name, err)
	}

	p.dataNamespace = "aptomi"
	if len(p.config.DataNamespace) > 0 {
		p.dataNamespace = clusterConfig.DataNamespace
	}
	if len(clusterConfig.DataNamespace) > 0 {
		p.dataNamespace = clusterConfig.DataNamespace
	}

	return nil
}
