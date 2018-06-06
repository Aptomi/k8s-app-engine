package k8sraw

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/k8s"
	"github.com/Aptomi/aptomi/pkg/util/sync"
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

func (p *Plugin) init() error {
	return p.once.Do(func() error {
		err := p.kube.Init()
		if err != nil {
			return err
		}

		err = p.parseClusterConfig()
		if err != nil {
			return err
		}

		kubeClient, err := p.kube.NewClient()
		if err != nil {
			return err
		}

		return p.kube.EnsureNamespace(kubeClient, p.dataNamespace)
	})
}

// Cleanup implements cleanup phase for the k8s raw plugin
func (p *Plugin) Cleanup() error {
	return nil
}

// Create implements creation of a new component instance in the cloud by deploying raw k8s objects
func (p *Plugin) Create(invocation *plugin.CodePluginInvocationParams) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	namespace := invocation.PluginParams[plugin.ParamTargetSuffix]
	if len(namespace) <= 0 {
		namespace = p.kube.DefaultNamespace
	}

	targetManifest, ok := invocation.Params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.kube.NewHelmKube(invocation.DeployName, invocation.EventLog)

	err = client.Create(namespace, strings.NewReader(targetManifest), 42, false)
	if err != nil {
		return err
	}

	return p.storeManifest(kubeClient, invocation.DeployName, targetManifest)
}

// Update implements update of an existing component instance in the cloud by updating raw k8s objects
func (p *Plugin) Update(invocation *plugin.CodePluginInvocationParams) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	namespace := invocation.PluginParams[plugin.ParamTargetSuffix]
	if len(namespace) <= 0 {
		namespace = p.kube.DefaultNamespace
	}

	currentManifest, err := p.loadManifest(kubeClient, invocation.DeployName)
	if err != nil {
		return err
	}

	targetManifest, ok := invocation.Params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.kube.NewHelmKube(invocation.DeployName, invocation.EventLog)

	err = client.Update(namespace, strings.NewReader(currentManifest), strings.NewReader(targetManifest), false, false, 42, false)
	if err != nil {
		return err
	}

	return p.storeManifest(kubeClient, invocation.DeployName, targetManifest)
}

// Destroy implements destruction of an existing component instance in the cloud by deleting raw k8s objects
func (p *Plugin) Destroy(invocation *plugin.CodePluginInvocationParams) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	namespace := invocation.PluginParams[plugin.ParamTargetSuffix]
	if len(namespace) <= 0 {
		namespace = p.kube.DefaultNamespace
	}

	deleteManifest, ok := invocation.Params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.kube.NewHelmKube(invocation.DeployName, invocation.EventLog)

	err = client.Delete(namespace, strings.NewReader(deleteManifest))
	if err != nil {
		return err
	}

	return p.deleteManifest(kubeClient, invocation.DeployName)
}

// Endpoints returns map from port type to url for all services of the deployed raw k8s objects
func (p *Plugin) Endpoints(invocation *plugin.CodePluginInvocationParams) (map[string]string, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	namespace := invocation.PluginParams[plugin.ParamTargetSuffix]
	if len(namespace) <= 0 {
		namespace = p.kube.DefaultNamespace
	}

	targetManifest, ok := invocation.Params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	return p.kube.EndpointsForManifests(namespace, invocation.DeployName, targetManifest, invocation.EventLog)
}

// Resources returns list of all resources (like services, config maps, etc.) deployed into the cluster by specified component instance
func (p *Plugin) Resources(invocation *plugin.CodePluginInvocationParams) (plugin.Resources, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	namespace := invocation.PluginParams[plugin.ParamTargetSuffix]
	if len(namespace) <= 0 {
		namespace = p.kube.DefaultNamespace
	}

	targetManifest, ok := invocation.Params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	return p.kube.ResourcesForManifest(namespace, invocation.DeployName, targetManifest, invocation.EventLog)
}

// Status returns readiness of all resources (like services, config maps, etc.) deployed into the cluster by specified component instance
func (p *Plugin) Status(invocation *plugin.CodePluginInvocationParams) (bool, error) {
	err := p.init()
	if err != nil {
		return false, err
	}

	namespace := invocation.PluginParams[plugin.ParamTargetSuffix]
	if len(namespace) <= 0 {
		namespace = p.kube.DefaultNamespace
	}

	targetManifest, ok := invocation.Params["manifest"].(string)
	if !ok {
		return false, fmt.Errorf("manifest is a mandatory parameter")
	}

	return p.kube.ReadinessStatusForManifest(namespace, invocation.DeployName, targetManifest, invocation.EventLog)
}
