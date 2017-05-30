package slinga

import (
	"fmt"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	log "github.com/Sirupsen/logrus"
)

type KubeClient struct {
	cluster *Cluster
	tillerHost string
	tillerTunnel *kube.Tunnel
}

func NewKubeClient(cluster *Cluster) *KubeClient {
	client := &KubeClient{cluster, "", nil}

	err := client.setupTillerConnection()
	if err != nil {
		debug.WithFields(log.Fields{
			"error": err,
		}).Fatal("Can't setup tiller connection")
	}

	return client
}


// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func (kubeClient *KubeClient) getKubeClient() (*restclient.Config, *internalclientset.Clientset, error) {
	kubeContext, ok := kubeClient.cluster.Metadata["kubeContext"]
	if !ok {
		debug.WithFields(log.Fields{
			"cluster": kubeClient.cluster,
		}).Fatal("Kube context should be specified for k8s cluster")
	}
	config, err := kube.GetConfig(kubeContext).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get kubernetes config for context '%s': %s", kubeContext, err)
	}
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get kubernetes client: %s", err)
	}
	return config, client, nil
}

func (kubeClient *KubeClient) setupTillerConnection() error {
	if tillerHost, ok := kubeClient.cluster.Metadata["tillerHost"]; ok {
		kubeClient.tillerHost = tillerHost
		return nil
	}

	config, client, err := kubeClient.getKubeClient()
	if err != nil {
		return err
	}

	tillerNamespace, ok := kubeClient.cluster.Metadata["tillerNamespace"]
	if !ok {
		tillerNamespace = "kube-system"
	}
	tunnel, err := portforwarder.New(tillerNamespace, client, config)
	if err != nil {
		return err
	}

	kubeClient.tillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)

	debug.WithFields(log.Fields{
		"port": tunnel.Local,
	}).Info("Created k8s tunnel using local port")

	return nil
}

func (kubeClient *KubeClient) Cleanup(){
	if kubeClient.tillerTunnel != nil {
		kubeClient.tillerTunnel.Close()
	}
}

