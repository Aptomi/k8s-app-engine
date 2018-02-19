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
	once    sync.Init
	cluster *lang.Cluster
	kube    *k8s.Plugin
}

// New returns new instance of the Kubernetes Raw code (objects) plugin for specified Kubernetes cluster plugin and plugins config
func New(clusterPlugin plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
	kubePlugin, ok := clusterPlugin.(*k8s.Plugin)
	if !ok {
		return nil, fmt.Errorf("k8s cluster plugin expected for k8sraw code plugin creation but received: %T", clusterPlugin)
	}

	return &Plugin{
		cluster: kubePlugin.Cluster,
		kube:    kubePlugin,
	}, nil
}

func (plugin *Plugin) init() error {
	return plugin.once.Do(func() error {
		return plugin.kube.Init()
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

	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("content is a mandatory parameter")
	}

	client := plugin.prepareClient(eventLog, deployName)

	return client.Create(plugin.kube.Namespace, strings.NewReader(content), 42, false)
}

func (plugin *Plugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := plugin.init()
	if err != nil {
		return err
	}

	// todo: implement me
	return fmt.Errorf("kubernetes raw code plugin doesn't support updates yet")
}

func (plugin *Plugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := plugin.init()
	if err != nil {
		return err
	}

	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("content is a mandatory parameter")
	}

	client := plugin.prepareClient(eventLog, deployName)

	return client.Delete(plugin.kube.Namespace, strings.NewReader(content))
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
