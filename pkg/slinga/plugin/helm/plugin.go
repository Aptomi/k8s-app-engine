package helm

import (
	"github.com/Aptomi/aptomi/pkg/slinga/config"
	"sync"
)

// Plugin uses Helm for deployment of apps on kubernetes
type Plugin struct {
	cache *sync.Map
	cfg   config.Helm
}

// NewPlugin creates a new helm plugin
func NewPlugin(cfg config.Helm) *Plugin {
	return &Plugin{
		cache: new(sync.Map),
		cfg:   cfg,
	}
}
