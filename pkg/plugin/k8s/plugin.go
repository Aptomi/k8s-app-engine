package k8s

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/util/sync"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Plugin represents Kubernetes cluster plugin
type Plugin struct {
	once             sync.Init
	config           config.K8s
	Cluster          *lang.Cluster
	RestConfig       *rest.Config
	ClientConfig     clientcmd.ClientConfig
	DefaultNamespace string
	ExternalAddress  string
}

var _ plugin.ClusterPlugin = &Plugin{}

// New creates new instance of the Kubernetes cluster plugin for specified Cluster and plugins config
func New(cluster *lang.Cluster, cfg config.Plugins) (plugin.ClusterPlugin, error) {
	return &Plugin{
		config:  cfg.K8s,
		Cluster: cluster,
	}, nil
}

// Validate checks Kubernetes cluster by connecting to it and ensuring default namespace
func (p *Plugin) Validate() error {
	err := p.Init()
	if err != nil {
		return err
	}

	client, err := p.NewClient()
	if err != nil {
		return err
	}

	return p.EnsureNamespace(client, p.DefaultNamespace)
}

// Init parses Kubernetes cluster config and retrieves external address for Kubernetes cluster
func (p *Plugin) Init() error {
	return p.once.Do(func() error {
		err := p.parseClusterConfig()
		if err != nil {
			return err
		}

		p.ExternalAddress, err = p.getExternalAddress()
		return err
	})
}

// Cleanup intended to run cleanup operations for plugin, but it's not used in Kubernetes cluster plugin
func (p *Plugin) Cleanup() error {
	// no cleanup needed
	return nil
}
