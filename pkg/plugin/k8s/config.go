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

func (p *Plugin) parseClusterConfig() error {
	cluster := p.Cluster

	clusterConfig := &ClusterConfig{}
	err := p.Cluster.ParseConfigInto(clusterConfig)
	if err != nil {
		return fmt.Errorf("error while parsing kubernetes specific config of cluster %s: %s", p.Cluster.Name, err)
	}

	if clusterConfig.Local && clusterConfig.KubeConfig != nil {
		return fmt.Errorf("kube-config can't be specified when using local type in cluster: %s", cluster.Name)
	}

	if clusterConfig.KubeConfig != nil {
		p.RestConfig, p.ClientConfig, p.Namespace, err = initKubeConfig(clusterConfig, cluster)
	} else {
		p.RestConfig, p.ClientConfig, err = initLocalKubeConfig(cluster)
	}
	if err != nil {
		return err
	}

	if p.config.Timeout == 0 {
		p.config.Timeout = 10 * time.Second
	}
	p.RestConfig.Timeout = p.config.Timeout

	if len(clusterConfig.Namespace) > 0 {
		p.Namespace = clusterConfig.Namespace
	}
	if len(p.Namespace) == 0 {
		p.Namespace = "default"
	}

	return nil
}

func initKubeConfig(config *ClusterConfig, cluster *lang.Cluster) (*rest.Config, clientcmd.ClientConfig, string, error) {
	var data []byte
	if strData, ok := config.KubeConfig.(string); ok {
		data = []byte(strData)
	} else {
		yamlData, err := yaml.Marshal(config.KubeConfig)
		if err != nil {
			return nil, nil, "", fmt.Errorf("error while marshaling kube config into bytes: %s", err)
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
		return nil, nil, "", fmt.Errorf("error while getting raw kube config for cluster %s: %s", cluster.Name, err)
	}

	if len(config.Context) == 0 && len(rawConf.CurrentContext) == 0 {
		return nil, nil, "", fmt.Errorf("context for cluster %s should be explicitly defined (context in cluster config or current-context in kubeconfig)", cluster.Name)
	}

	clientConf, err := conf.ClientConfig()
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not get kubernetes config for cluster %s: %s", cluster.Name, err)
	}

	if namespace, _, nsErr := conf.Namespace(); nsErr == nil && len(namespace) > 0 {
		return clientConf, conf, namespace, nil
	}

	return clientConf, conf, "", nil
}

func initLocalKubeConfig(cluster *lang.Cluster) (*rest.Config, clientcmd.ClientConfig, error) {
	rules := &clientcmd.ClientConfigLoadingRules{}
	overrides := &clientcmd.ConfigOverrides{}
	conf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	clientConf, err := conf.ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for cluster %s: %s", cluster.Name, err)
	}

	return clientConf, conf, nil
}
