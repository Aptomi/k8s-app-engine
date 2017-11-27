package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
)

// Config represents K8s/Helm plugin configuration
type Config struct {
	Context         string
	Namespace       string
	TillerNamespace string
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

	data, err := yaml.Marshal(config.KubeConfig)
	if err != nil {
		return fmt.Errorf("error while marshaling kube config into bytes: %s", err)
	}

	// todo make sure temp file removed after kube config created
	kubeConfigFile := util.WriteTempFile(fmt.Sprintf("cluster-config"), data)

	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigFile}

	overrides := &clientcmd.ConfigOverrides{}
	if len(config.Context) > 0 {
		overrides.CurrentContext = config.Context
	}

	conf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	rawConf, err := conf.RawConfig()
	if err != nil {
		return fmt.Errorf("error while getting raw kube config for cluster %s: %s", cluster.Name, err)
	}

	if len(config.Context) == 0 && len(rawConf.CurrentContext) == 0 {
		return fmt.Errorf("context for cluster %s should be explicitly defined (context in cluster config or current-context in kubeconfig)", cluster.Name)
	}

	clientConf, err := conf.ClientConfig()
	if err != nil {
		return fmt.Errorf("could not get kubernetes config for cluster %s: %s", cache.cluster.Name, err)
	}
	cache.kubeConfig = clientConf

	if len(config.Namespace) > 0 {
		cache.namespace = config.Namespace
	} else if namespace, _, nsErr := conf.Namespace(); nsErr == nil && len(namespace) > 0 {
		cache.namespace = namespace
	} else {
		cache.namespace = "default"
	}

	cache.tillerNamespace = "kube-system"
	if len(config.TillerNamespace) > 0 {
		cache.tillerNamespace = config.TillerNamespace
	}

	return nil
}
