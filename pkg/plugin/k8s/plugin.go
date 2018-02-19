package k8s

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"k8s.io/client-go/rest"
	"sync"
)

// Plugin represents Kubernetes cluster plugin
type Plugin struct {
	once            sync.Once
	config          config.Kube
	Cluster         *lang.Cluster
	KubeConfig      *rest.Config
	Namespace       string
	ExternalAddress string
}

var _ plugin.ClusterPlugin = &Plugin{}

// New creates new instance of the Kubernetes cluster plugin for specified Cluster and plugins config
func New(cluster *lang.Cluster, cfg config.Plugins) (plugin.ClusterPlugin, error) {
	return &Plugin{
		config:  cfg.Kube,
		Cluster: cluster,
	}, nil
}

// Validate checks Kubernetes cluster by connecting to it and ensuring configured namespace
func (plugin *Plugin) Validate() error {
	err := plugin.Init()
	if err != nil {
		return err
	}

	client, err := plugin.NewClient()
	if err != nil {
		return err
	}

	return plugin.EnsureNamespace(client, plugin.Namespace)
}

// Init parses Kubernetes cluster config and retrieves external address for Kubernetes cluster
func (plugin *Plugin) Init() (err error) {
	plugin.once.Do(func() {
		err = plugin.parseClusterConfig()
		if err != nil {
			return
		}

		plugin.ExternalAddress, err = plugin.getExternalAddress()
		if err != nil {
			return
		}
	})
	return
}

// Cleanup intended to run cleanup operations for plugin, but it's not used in Kubernetes cluster plugin
func (plugin *Plugin) Cleanup() error {
	// no cleanup needed
	return nil
}
