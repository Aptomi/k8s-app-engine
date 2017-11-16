package helm

import (
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"k8s.io/helm/pkg/kube"
	"sync"
)

type clusterCache struct {
	cluster             *lang.Cluster
	lock                sync.Mutex   // all caching ops should use this lock
	kubeExternalAddress string       // kube external address
	tillerTunnel        *kube.Tunnel // tunnel for accessing tiller
	tillerHost          string       // local proxy address when connection established
	istioSvc            string       // istio svc name
}

func (plugin *Plugin) getClusterCache(cluster *lang.Cluster, eventLog *event.Log) (*clusterCache, error) {
	rawCache, loaded := plugin.cache.LoadOrStore(cluster.Name, new(clusterCache))
	cache := rawCache.(*clusterCache)
	if !loaded {
		cache.cluster = cluster
		err := cache.init(eventLog)
		if err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func (cache *clusterCache) init(eventLog *event.Log) error {
	return cache.ensureTillerTunnel(eventLog)
}
