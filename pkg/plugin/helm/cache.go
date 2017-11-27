package helm

import (
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/kube"
	"sync"
)

type clusterCache struct {
	cluster         *lang.Cluster
	config          *Config
	lock            sync.Mutex // all caching ops should use this lock
	kubeConfig      *rest.Config
	namespace       string
	tillerNamespace string
	externalAddress string       // kube external address
	tillerTunnel    *kube.Tunnel // tunnel for accessing tiller
	tillerHost      string       // local proxy address when connection established
	istioSvc        string       // istio svc name
}

func (plugin *Plugin) getClusterCache(cluster *lang.Cluster, eventLog *event.Log) (*clusterCache, error) {
	rawCache, loaded := plugin.cache.LoadOrStore(cluster.Name, new(clusterCache))
	cache := rawCache.(*clusterCache)
	if !loaded {
		err := cache.init(cluster, eventLog)
		if err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func (cache *clusterCache) init(cluster *lang.Cluster, eventLog *event.Log) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	err := cache.initConfig(cluster)
	if err != nil {
		return err
	}

	return cache.ensureTillerTunnel(eventLog)
}
