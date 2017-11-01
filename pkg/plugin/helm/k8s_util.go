package helm

import (
	"errors"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm/portforwarder"
)

func (cache *clusterCache) setupTillerConnection(cluster *lang.Cluster, eventLog *event.Log) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if len(cache.tillerHost) > 0 {
		// todo(slukjanov): verify that tunnel is still alive??
		// connection already set up, skip
		return nil
	}

	config, client, err := cache.newKubeClient(cluster)
	if err != nil {
		return err
	}

	tillerNamespace := cluster.Config.TillerNamespace
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	tunnel, err := portforwarder.New(tillerNamespace, client, config)
	if err != nil {
		return err
	}

	cache.tillerTunnel = tunnel
	cache.tillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)

	eventLog.WithFields(event.Fields{}).Debugf("Created k8s tunnel using local port: %d", tunnel.Local)

	return nil
}

func (cache *clusterCache) getK8sClientConfig(cluster *lang.Cluster) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	/*
		if len(cluster.Config.KubeConfig) > 0 {
			rules = &clientcmd.ClientConfigLoadingRules{ExplicitPath: cluster.Config.KubeConfig}
		}
	*/

	kubeContext := cluster.Config.KubeContext
	overrides := &clientcmd.ConfigOverrides{}
	if len(kubeContext) > 0 {
		overrides.CurrentContext = kubeContext
	}
	conf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	clientConf, err := conf.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get kubernetes config for cluster %s: %s", cluster.Name, err)
	}

	return clientConf, nil
}

func (cache *clusterCache) newKubeClient(cluster *lang.Cluster) (*rest.Config, kubernetes.Interface, error) {
	conf, err := cache.getK8sClientConfig(cluster)
	if err != nil {
		return nil, nil, err
	}

	client, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}

	return conf, client, nil
}

func (cache *clusterCache) getKubeExternalAddress(cluster *lang.Cluster, eventLog *event.Log) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if len(cache.kubeExternalAddress) > 0 {
		return cache.kubeExternalAddress, nil
	}

	_, client, err := cache.newKubeClient(cluster)
	if err != nil {
		return "", fmt.Errorf("error while creating k8s client to cluster %s: %s", cluster.Name, err)
	}

	nodes, err := client.CoreV1().Nodes().List(meta.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found for k8s cluster %s, it's critical eror", cluster.Name)
	}

	returnFirst := func(addrType api.NodeAddressType) string {
		for _, node := range nodes.Items {
			for _, addr := range node.Status.Addresses {
				if addr.Type == addrType {
					return addr.Address
				}
			}
		}
		return ""
	}

	addr := returnFirst(api.NodeExternalIP)
	if addr == "" {
		addr = returnFirst(api.NodeInternalIP)
	}
	if addr == "" {
		addr = returnFirst(api.NodeHostName)
	}
	if addr == "" {
		return "", errors.New("couldn't find external IP for cluster")
	}

	cache.kubeExternalAddress = addr

	return addr, nil
}
