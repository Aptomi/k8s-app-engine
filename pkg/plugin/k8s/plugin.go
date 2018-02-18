package k8s

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"k8s.io/client-go/rest"
	"sync"
)

type Plugin struct {
	once            sync.Once
	config          config.Kube
	Cluster         *lang.Cluster
	KubeConfig      *rest.Config
	Namespace       string
	ExternalAddress string
}

var _ plugin.ClusterPlugin = &Plugin{}

func New(cluster *lang.Cluster, cfg config.Plugins) (plugin.ClusterPlugin, error) {
	return &Plugin{
		config:  cfg.Kube,
		Cluster: cluster,
	}, nil
}

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

func (plugin *Plugin) Cleanup() error {
	// no cleanup needed
	return nil
}
