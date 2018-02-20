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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/kube"
	"strings"
)

// Plugin represents Kubernetes Raw code plugin that supports deploying specified k8s objects into the cluster
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

// Cleanup implements cleanup phase for the k8s raw plugin
func (plugin *Plugin) Cleanup() error {
	return nil
}

// Create implements creation of a new component instance in the cloud by deploying raw k8s objects
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

// Update implements update of an existing component instance in the cloud by updating raw k8s objects
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

// Destroy implements destruction of an existing component instance in the cloud by deleting raw k8s objects
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

// Endpoints returns map from port type to url for all services of the deployed raw k8s objects
func (plugin *Plugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	err := plugin.init()
	if err != nil {
		return nil, err
	}

	kubeClient, err := plugin.kube.NewClient()
	if err != nil {
		return nil, err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	client := plugin.prepareClient(eventLog, deployName)

	infos, err := client.BuildUnstructured(plugin.kube.Namespace, strings.NewReader(targetManifest))
	if err != nil {
		return nil, err
	}

	endpoints := make(map[string]string)

	for _, info := range infos {
		if info.Mapping.GroupVersionKind.Kind == "Service" {
			service, getErr := kubeClient.CoreV1().Services(plugin.kube.Namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return nil, getErr
			}

			plugin.kube.AddEndpointsFromService(service, endpoints)
		}
	}

	return endpoints, nil
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
