package k8sraw

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/k8s"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Aptomi/aptomi/pkg/util/sync"
	"k8s.io/helm/pkg/kube"
	"strings"
)

type Plugin struct {
	once          sync.Init
	cluster       *lang.Cluster
	config        config.K8sRaw
	kube          *k8s.Plugin
	dataNamespace string
}

// New returns new instance of the Kubernetes Raw code (objects) plugin for specified Kubernetes cluster plugin and plugins config
func New(clusterPlugin plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
	kubePlugin, ok := clusterPlugin.(*k8s.Plugin)
	if !ok {
		return nil, fmt.Errorf("k8s cluster plugin expected for k8sraw code plugin creation but received: %T", clusterPlugin)
	}

	return &Plugin{
		cluster: kubePlugin.Cluster,
		config:  cfg.K8sRaw,
		kube:    kubePlugin,
	}, nil
}

func (plugin *Plugin) init() error {
	return plugin.once.Do(func() error {
		err := plugin.kube.Init()
		if err != nil {
			return err
		}

		err = plugin.parseClusterConfig()
		if err != nil {
			return err
		}

		kubeClient, err := plugin.kube.NewClient()
		if err != nil {
			return err
		}

		return plugin.kube.EnsureNamespace(kubeClient, plugin.dataNamespace)
	})
}

func (plugin *Plugin) Cleanup() error {
	return nil
}

func (plugin *Plugin) Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := plugin.init()
	if err != nil {
		return err
	}

	kubeClient, err := plugin.kube.NewClient()
	if err != nil {
		return err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := plugin.prepareClient(eventLog, deployName)

	err = client.Create(plugin.kube.Namespace, strings.NewReader(targetManifest), 42, false)
	if err != nil {
		return err
	}

	return plugin.storeManifest(kubeClient, deployName, targetManifest)
}

func (plugin *Plugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := plugin.init()
	if err != nil {
		return err
	}

	kubeClient, err := plugin.kube.NewClient()
	if err != nil {
		return err
	}

	currentManifest, err := plugin.loadManifest(kubeClient, deployName)
	if err != nil {
		return err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := plugin.prepareClient(eventLog, deployName)

	err = client.Update(plugin.kube.Namespace, strings.NewReader(currentManifest), strings.NewReader(targetManifest), false, false, 42, false)
	if err != nil {
		return err
	}

	return plugin.storeManifest(kubeClient, deployName, targetManifest)
}

func (plugin *Plugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := plugin.init()
	if err != nil {
		return err
	}

	kubeClient, err := plugin.kube.NewClient()
	if err != nil {
		return err
	}

	deleteManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := plugin.prepareClient(eventLog, deployName)

	err = client.Delete(plugin.kube.Namespace, strings.NewReader(deleteManifest))
	if err != nil {
		return err
	}

	return plugin.deleteManifest(kubeClient, deployName)
}

func (plugin *Plugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	err := plugin.init()
	if err != nil {
		return nil, err
	}

	// todo: implement me
	return make(map[string]string), nil
}

func (plugin *Plugin) prepareClient(eventLog *event.Log, deployName string) *kube.Client {
	client := kube.New(plugin.kube.ClientConfig)
	client.Log = func(format string, args ...interface{}) {
		eventLog.WithFields(event.Fields{
			"deployName": deployName,
		}).Debugf(fmt.Sprintf("[instance: %s] ", deployName)+format, args...)
	}

	return client
}
