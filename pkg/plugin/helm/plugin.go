package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/k8s"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Aptomi/aptomi/pkg/util/sync"
	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/kube"
	"strings"
)

// Plugin represents Helm code plugin for Kubernetes cluster
type Plugin struct {
	once            sync.Init
	cluster         *lang.Cluster
	config          config.Helm
	kube            *k8s.Plugin
	tillerNamespace string       // namespace for tiller
	tillerTunnel    *kube.Tunnel // tunnel for accessing tiller
	tillerHost      string       // local proxy address when connection established
}

var _ plugin.CodePlugin = &Plugin{}

// New returns new instance of the Helm code plugin for specified Kubernetes cluster plugin and plugins config
func New(clusterPlugin plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
	kubePlugin, ok := clusterPlugin.(*k8s.Plugin)
	if !ok {
		return nil, fmt.Errorf("k8s cluster plugin expected for helm code plugin creation but received: %T", clusterPlugin)
	}

	return &Plugin{
		config:  cfg.Helm,
		kube:    kubePlugin,
		cluster: kubePlugin.Cluster,
	}, nil
}

func (p *Plugin) init(eventLog *event.Log) error {
	return p.once.Do(func() error {
		err := p.kube.Init()
		if err != nil {
			return err
		}

		err = p.parseClusterConfig()
		if err != nil {
			return err
		}

		// todo(slukjanov): we should probably verify tunnel each time we need it
		return p.ensureTillerTunnel(eventLog)
	})
}

// Cleanup implements cleanup phase for the Helm plugin. It closes cached Tiller tunnel.
func (p *Plugin) Cleanup() error {
	if p.tillerTunnel != nil {
		p.tillerTunnel.Close()
	}

	return nil
}

// Create implements creation of a new component instance in the cloud by deploying a Helm chart
func (p *Plugin) Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	return p.createOrUpdate(deployName, params, eventLog, true)
}

// Update implements update of an existing component instance in the cloud by updating parameters of a helm chart
func (p *Plugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	return p.createOrUpdate(deployName, params, eventLog, false)
}

func (p *Plugin) createOrUpdate(deployName string, params util.NestedParameterMap, eventLog *event.Log, create bool) error {
	err := p.init(eventLog)
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	err = p.kube.EnsureNamespace(kubeClient, p.kube.Namespace)
	if err != nil {
		return err
	}

	releaseName := getReleaseName(deployName)
	chartRepo, chartName, chartVersion, err := getHelmReleaseInfo(params)
	if err != nil {
		return err
	}

	helmClient, err := p.newClient()
	if err != nil {
		return err
	}

	chartPath, err := p.fetchChart(chartRepo, chartName, chartVersion)
	if err != nil {
		return err
	}

	helmParams, err := yaml.Marshal(params)
	if err != nil {
		return err
	}

	currRelease, err := helmClient.ReleaseContent(releaseName)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("error while looking for Helm release %s: %s", releaseName, err)
	}

	cluster := p.cluster
	if create {
		if currRelease != nil {
			// If a release already exists, let's just go ahead and update it
			eventLog.WithFields(event.Fields{}).Infof("Release '%s' already exists. Updating it", releaseName)
		} else {
			eventLog.WithFields(event.Fields{
				"release": releaseName,
				"chart":   chartName,
				"path":    chartPath,
				"params":  string(helmParams),
			}).Infof("Installing Helm release '%s', chart '%s', cluster: '%s'", releaseName, chartName, cluster.Name)

			_, err = helmClient.InstallRelease(
				chartPath,
				p.kube.Namespace,
				helm.ReleaseName(releaseName),
				helm.ValueOverrides(helmParams),
				helm.InstallReuseName(true),
				helm.InstallTimeout(int64(p.config.Timeout)),
			)

			return err
		}
	}

	eventLog.WithFields(event.Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(helmParams),
	}).Infof("Updating Helm release '%s', chart '%s', cluster: '%s'", releaseName, chartName, cluster.Name)

	status, err := helmClient.ReleaseStatus(releaseName)
	if err != nil {
		return fmt.Errorf("error while getting status of current release %s: %s", releaseName, err)
	}
	if status.Namespace != p.kube.Namespace {
		return fmt.Errorf("it's not allowed to change namespace of the release %s (was %s, requested %s)", releaseName, status.Namespace, p.kube.Namespace)
	}

	newRelease, err := helmClient.UpdateRelease(
		releaseName,
		chartPath,
		helm.UpdateValueOverrides(helmParams),
		helm.UpgradeTimeout(int64(p.config.Timeout)),
	)
	if err != nil {
		return err
	}

	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(currRelease.Release.Manifest),
		B:        difflib.SplitLines(newRelease.Release.Manifest),
		FromFile: "Previous",
		ToFile:   "Current",
		Context:  3,
	})
	if err != nil {
		return fmt.Errorf("error while calculating diff between chart manifests for Helm release '%s', chart '%s', cluster: '%s'", releaseName, chartName, cluster.Name)
	}

	if len(diff) == 0 {
		diff = "without changes"
	} else {
		diff = "with diff: \n\n" + diff
	}

	eventLog.WithFields(event.Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(helmParams),
	}).Debugf("Updated Helm release '%s', chart '%s', cluster: '%s' %s", releaseName, chartName, cluster.Name, diff)

	return err
}

// Destroy implements destruction of an existing component instance in the cloud by running "helm delete" on the corresponding helm chart
func (p *Plugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init(eventLog)
	if err != nil {
		return err
	}

	releaseName := getReleaseName(deployName)

	helmClient, err := p.newClient()
	if err != nil {
		return err
	}

	eventLog.WithFields(event.Fields{
		"release": releaseName,
	}).Infof("Deleting Helm release '%s'", releaseName)

	_, err = helmClient.DeleteRelease(
		releaseName,
		helm.DeletePurge(true),
		helm.DeleteTimeout(int64(p.config.Timeout)),
	)
	return err
}

// Endpoints returns map from port type to url for all services of the current chart
func (p *Plugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	err := p.init(eventLog)
	if err != nil {
		return nil, err
	}

	helmClient, err := p.newClient()
	if err != nil {
		return nil, err
	}

	releaseName := getReleaseName(deployName)

	currRelease, err := helmClient.ReleaseContent(releaseName)
	if err != nil {
		return nil, fmt.Errorf("error while looking for Helm release %s: %s", releaseName, err)
	}

	return p.kube.EndpointsForManifests(deployName, currRelease.Release.Manifest, eventLog)
}

// Resources returns list of all resources (like services, config maps, etc.) deployed into the cluster by specified component instance
func (p *Plugin) Resources(deployName string, params util.NestedParameterMap, eventLog *event.Log) (plugin.Resources, error) {
	err := p.init(eventLog)
	if err != nil {
		return nil, err
	}

	helmClient, err := p.newClient()
	if err != nil {
		return nil, err
	}

	releaseName := getReleaseName(deployName)

	currRelease, err := helmClient.ReleaseContent(releaseName)
	if err != nil {
		return nil, fmt.Errorf("error while looking for Helm release %s: %s", releaseName, err)
	}

	return p.kube.ResourcesForManifest(deployName, currRelease.Release.Manifest, eventLog)
}

// Status returns readiness of all resources (like services, config maps, etc.) deployed into the cluster by specified component instance
func (p *Plugin) Status(deployName string, params util.NestedParameterMap, eventLog *event.Log) (bool, error) {
	err := p.init(eventLog)
	if err != nil {
		return false, err
	}

	helmClient, err := p.newClient()
	if err != nil {
		return false, err
	}

	releaseName := getReleaseName(deployName)

	_, err = helmClient.ReleaseContent(releaseName)
	if err != nil {
		return false, fmt.Errorf("error while looking for Helm release %s: %s", releaseName, err)
	}

	return false, nil
}
