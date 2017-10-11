package helm

import (
	"errors"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
)

func (cache *clusterCache) setupTillerConnection(cluster *lang.Cluster, eventLog *event.Log) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if len(cache.tillerHost) > 0 {
		// todo(slukjanov): verify that tunnel is still alive??
		// connection already set up, skip
		return nil
	}

	config, client, err := cache.newKubeClient(cluster, eventLog)
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

func (cache *clusterCache) newKubeClient(cluster *lang.Cluster, eventLog *event.Log) (*restclient.Config, *internalclientset.Clientset, error) {
	// todo(slukjanov): cache kube client config?
	kubeContext := cluster.Config.KubeContext
	config, err := kube.GetConfig(kubeContext).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get kubernetes config for context %s: %s", kubeContext, err)
	}
	// todo(slukjanov): could we cache client?
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get kubernetes client: %s", err)
	}
	return config, client, nil
}

func (cache *clusterCache) getKubeExternalAddress(cluster *lang.Cluster, eventLog *event.Log) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if len(cache.kubeExternalAddress) > 0 {
		return cache.kubeExternalAddress, nil
	}

	_, client, err := cache.newKubeClient(cluster, eventLog)
	if err != nil {
		return "", fmt.Errorf("Error while creating k8s client to cluster %s: %s", cluster.Name, err)
	}

	nodes, err := client.Nodes().List(api.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("No nodes found for k8s cluster %s, it's critical eror", cluster.Name)
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
		// TODO: this will be removed in 1.7
		addr = returnFirst(api.NodeLegacyHostIP)
	}
	if addr == "" {
		addr = returnFirst(api.NodeInternalIP)
	}
	if addr == "" {
		return "", errors.New("Couldn't find external IP for cluster")
	}

	cache.kubeExternalAddress = addr

	return addr, nil
}
