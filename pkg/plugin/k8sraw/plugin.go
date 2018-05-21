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
func (p *Plugin) Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.kube.NewHelmKube(deployName, eventLog)

	err = client.Create(p.kube.Namespace, strings.NewReader(targetManifest), 42, false)
	if err != nil {
		return err
	}

	return p.storeManifest(kubeClient, deployName, targetManifest)
}

// Update implements update of an existing component instance in the cloud by updating raw k8s objects
func (p *Plugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	currentManifest, err := p.loadManifest(kubeClient, deployName)
	if err != nil {
		return err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.kube.NewHelmKube(deployName, eventLog)

	err = client.Update(p.kube.Namespace, strings.NewReader(currentManifest), strings.NewReader(targetManifest), false, false, 42, false)
	if err != nil {
		return err
	}

	return p.storeManifest(kubeClient, deployName, targetManifest)
}

// Destroy implements destruction of an existing component instance in the cloud by deleting raw k8s objects
func (p *Plugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	deleteManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.kube.NewHelmKube(deployName, eventLog)

	err = client.Delete(p.kube.Namespace, strings.NewReader(deleteManifest))
	if err != nil {
		return err
	}

	return p.deleteManifest(kubeClient, deployName)
}

// Endpoints returns map from port type to url for all services of the deployed raw k8s objects
func (p *Plugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	return p.kube.EndpointsForManifests(deployName, targetManifest, eventLog)
}

// Resources returns list of all resources (like services, config maps, etc.) deployed into the cluster by specified component instance
func (p *Plugin) Resources(deployName string, params util.NestedParameterMap, eventLog *event.Log) (plugin.Resources, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	return p.kube.ResourcesForManifest(deployName, targetManifest, eventLog)
}

// Status returns readiness of all resources (like services, config maps, etc.) deployed into the cluster by specified component instance
func (p *Plugin) Status(deployName string, params util.NestedParameterMap, eventLog *event.Log) (bool, error) {
	err := p.init()
	if err != nil {
		return false, err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return false, fmt.Errorf("manifest is a mandatory parameter")
	}

	return p.kube.ReadinessStatusForManifest(deployName, targetManifest, eventLog)
}
