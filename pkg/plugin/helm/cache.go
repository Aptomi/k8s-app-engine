package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"k8s.io/helm/pkg/kube"
	"sync"
)

type clusterCache struct {
	lock                sync.Mutex   // all caching ops should use this lock
	tillerTunnel        *kube.Tunnel // tunnel for accessing tiller
	tillerHost          string       // local proxy address when connection established
	kubeExternalAddress string       // kube external address
	istioSvc            string       // istio svc name
}

func (p *Plugin) getCache(cluster *lang.Cluster, eventLog *event.Log) (*clusterCache, error) {
	cache, _ /*loaded*/ := p.cache.LoadOrStore(cluster.Name, new(clusterCache))
	c, ok := cache.(*clusterCache)
	if ok {
		err := c.setupTillerConnection(cluster, eventLog)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	panic(fmt.Sprintf("clusterCache expected in Plugin cache, but found: %v", c))
}
