package helm

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"sync"
)

// Plugin uses Helm for deployment of apps on kubernetes
type Plugin struct {
	cache *sync.Map
	cfg   config.Helm
}

var _ plugin.Plugin = &Plugin{}
var _ plugin.DeployPlugin = &Plugin{}
var _ plugin.PostProcessPlugin = &Plugin{}

// NewPlugin creates a new helm plugin
func NewPlugin(cfg config.Helm) *Plugin {
	return &Plugin{
		cache: new(sync.Map),
		cfg:   cfg,
	}
}
