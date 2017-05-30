package slinga

import (
	"sync"
	"errors"
)

var (
	clusterClients = make(map[string]ClusterClient)
	clusterClientsLock sync.Mutex
)

type ClusterClient interface {
	Cleanup()
}

func newClusterClient(cluster *Cluster) (ClusterClient, error) {
	switch cluster.Type {
	case "kubernetes":
		return NewKubeClient(cluster), nil
	default:
		return nil, errors.New("ClusterClient not found: " + cluster.Type)
	}
}

func (cluster *Cluster) Client() ClusterClient {
	clusterClientsLock.Lock()
	defer clusterClientsLock.Unlock()

	if client, ok := clusterClients[cluster.Name]; ok {
		return client
	}

	client, err := newClusterClient(cluster)
	if err != nil {
		// todo panic!
		panic(err)
	}
	clusterClients[cluster.Name] = client

	return client
}

// todo run cleanup somewhere to delete k8s tunnels
func CleanupClients() {
	clusterClientsLock.Lock()
	defer clusterClientsLock.Unlock()

	for key, client := range clusterClients {
		client.Cleanup()
		delete(clusterClients, key)
	}
}
