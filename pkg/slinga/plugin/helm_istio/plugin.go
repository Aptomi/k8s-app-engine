package helm_istio

import (
	"sync"
)

// HelmIstioPlugin is an executor that uses Helm for deployment of apps on kubernetes
type HelmIstioPlugin struct {
	cache *sync.Map
}

func NewHelmIstioPlugin() *HelmIstioPlugin {
	return &HelmIstioPlugin{
		cache: new(sync.Map),
	}
}
