package helm_istio

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"sync"
)

type clusterCache struct {
	lock                sync.Mutex // all caching ops should use this lock
	tillerHost          string     // store local proxy address when connection established
	kubeExternalAddress string     // store kube external address
	istioSvc            string     // store istio svc name
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
