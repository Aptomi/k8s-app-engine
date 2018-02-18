package k8s

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

// ClusterConfig represents Kubernetes cluster plugin configuration
type ClusterConfig struct {
	Namespace  string      `yaml:",omitempty"`
	Local      bool        `yaml:",omitempty"`
	Context    string      `yaml:",omitempty"`
	KubeConfig interface{} `yaml:",omitempty"` // it's just a kubeconfig, we don't need to parse it
}

func (plugin *Plugin) parseClusterConfig() error {
	cluster := plugin.Cluster

	clusterConfig := &ClusterConfig{}
	err := plugin.Cluster.ParseConfigInto(clusterConfig)
	if err != nil {
		err = fmt.Errorf("error while parsing kubernetes specific config of cluster %s: %s", plugin.Cluster.Name, err)
	}

	if clusterConfig.Local && clusterConfig.KubeConfig != nil {
		return fmt.Errorf("kube-config can't be specified when using local type in cluster: %s", cluster.Name)
	}

	if clusterConfig.KubeConfig != nil {
		plugin.KubeConfig, plugin.Namespace, err = initKubeConfig(clusterConfig, cluster)
	} else {
		plugin.KubeConfig, err = initLocalKubeConfig()
	}
	if err != nil {
		return err
	}

	if plugin.config.Timeout == 0 {
		plugin.config.Timeout = 10 * time.Second
	}
	plugin.KubeConfig.Timeout = plugin.config.Timeout

	if len(clusterConfig.Namespace) > 0 {
		plugin.Namespace = clusterConfig.Namespace
	}
	if len(plugin.Namespace) == 0 {
		plugin.Namespace = "default"
	}

	return nil
}

func initKubeConfig(config *ClusterConfig, cluster *lang.Cluster) (*rest.Config, string, error) {
	var data []byte
	if strData, ok := config.KubeConfig.(string); ok {
		data = []byte(strData)
	} else {
		yamlData, err := yaml.Marshal(config.KubeConfig)
		if err != nil {
			return nil, "", fmt.Errorf("error while marshaling kube config into bytes: %s", err)
		}
		data = yamlData
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
		return nil, "", fmt.Errorf("error while getting raw kube config for cluster %s: %s", cluster.Name, err)
	}

	if len(config.Context) == 0 && len(rawConf.CurrentContext) == 0 {
		return nil, "", fmt.Errorf("context for cluster %s should be explicitly defined (context in cluster config or current-context in kubeconfig)", cluster.Name)
	}

	clientConf, err := conf.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("could not get kubernetes config for cluster %s: %s", cluster.Name, err)
	}

	if namespace, _, nsErr := conf.Namespace(); nsErr == nil && len(namespace) > 0 {
		return clientConf, namespace, nil
	}

	return clientConf, "", nil
}

func initLocalKubeConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}
