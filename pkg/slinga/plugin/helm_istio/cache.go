package helm_istio

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
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

func (p *HelmIstioPlugin) getCache(cluster *lang.Cluster, eventLog *eventlog.EventLog) (*clusterCache, error) {
	cache, _ /*loaded*/ := p.cache.LoadOrStore(cluster.Namespace, new(clusterCache))
	if c, ok := cache.(*clusterCache); ok {
		err := c.setupTillerConnection(cluster, eventLog)
		if err != nil {
			return nil, err
		}
		return c, nil
	} else {
		panic(fmt.Sprintf("clusterCache expected in HelmIstioPlugin cache, but found: %v", c))
	}
}

func (p *HelmIstioPlugin) Cleanup() error {
	var err error
	p.cache.Range(func(key, value interface{}) bool {
		if c, ok := value.(*clusterCache); ok {
			c.tillerTunnel.Close()
		} else {
			panic(fmt.Sprintf("clusterCache expected in HelmIstioPlugin cache, but found: %v", c))
		}
		return true
	})
	return err
}
