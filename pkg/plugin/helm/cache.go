package helm

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/kube"
	"sync"
)

type clusterCache struct {
	pluginConfig    config.Helm
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
		err := cache.init(plugin.cfg, cluster, eventLog)
		if err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func (cache *clusterCache) init(pluginConfig config.Helm, cluster *lang.Cluster, eventLog *event.Log) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.pluginConfig = pluginConfig
	err := cache.initConfig(cluster)
	if err != nil {
		return err
	}

	return cache.ensureTillerTunnel(eventLog)
}
