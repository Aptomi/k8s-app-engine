package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
)

// Config represents K8s/Helm plugin configuration
type Config struct {
	Namespace       string
	TillerNamespace string
	Local           bool
	Context         string
	KubeConfig      interface{} // it's just a kubeconfig, we don't need to parse it
}

func (cache *clusterCache) initConfig(cluster *lang.Cluster) error {
	cache.cluster = cluster

	config := &Config{}
	cache.config = config

	err := cluster.ParseConfigInto(config)
	if err != nil {
		return fmt.Errorf("error while parsing Helm plugin specific cluster config: %s", err)
	}

	if config.Local && config.KubeConfig != nil {
		return fmt.Errorf("kube-config can't be specified when using local type in cluster: %s", cluster.Name)
	}

	if config.KubeConfig != nil {
		cache.kubeConfig, cache.namespace, err = initKubeConfig(config, cluster)
	} else {
		cache.kubeConfig, err = initLocalKubeConfig()
	}
	if err != nil {
		return err
	}

	if len(config.Namespace) > 0 {
		cache.namespace = config.Namespace
	}
	if len(cache.namespace) == 0 {
		cache.namespace = "default"
	}

	cache.tillerNamespace = "kube-system"
	if len(config.TillerNamespace) > 0 {
		cache.tillerNamespace = config.TillerNamespace
	}

	return nil
}
