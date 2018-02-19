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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	namespace       string       // namespace to deploy app to
	tillerNamespace string       // namespace for tiller
	tillerTunnel    *kube.Tunnel // tunnel for accessing tiller
	tillerHost      string       // local proxy address when connection established
}

var _ plugin.CodePlugin = &Plugin{}

// New returns new instance of the Helm code plugin for specified Kubernetes cluster plugin and plugins config
func New(clusterPlugin plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
	kubePlugin, ok := clusterPlugin.(*k8s.Plugin)
	if !ok {
		return nil, fmt.Errorf("kube cluster plugin expected for helm code plugin creation but received: %T", clusterPlugin)
	}

	return &Plugin{
		config:    cfg.Helm,
		kube:      kubePlugin,
		cluster:   kubePlugin.Cluster,
		namespace: kubePlugin.Namespace,
	}, nil
}

func (plugin *Plugin) init(eventLog *event.Log) error {
	return plugin.once.Do(func() error {
		err := plugin.kube.Init()
		if err != nil {
			return err
		}

		err = plugin.parseClusterConfig()
		if err != nil {
			return err
		}

		// todo(slukjanov): we should probably verify tunnel each time we need it
		return plugin.ensureTillerTunnel(eventLog)
	})
}

// Cleanup implements cleanup phase for the Helm plugin. It closes cached Tiller tunnel.
func (plugin *Plugin) Cleanup() error {
	if plugin.tillerTunnel != nil {
		plugin.tillerTunnel.Close()
	}

	return nil
}

// Create implements creation of a new component instance in the cloud by deploying a Helm chart
func (plugin *Plugin) Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	return plugin.createOrUpdate(deployName, params, eventLog, true)
}

// Update implements update of an existing component instance in the cloud by updating parameters of a helm chart
func (plugin *Plugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	return plugin.createOrUpdate(deployName, params, eventLog, false)
}

func (plugin *Plugin) createOrUpdate(deployName string, params util.NestedParameterMap, eventLog *event.Log, create bool) error {
	err := plugin.init(eventLog)
	if err != nil {
		return err
	}

	kubeClient, err := plugin.kube.NewClient()
	if err != nil {
		return err
	}

	err = plugin.kube.EnsureNamespace(kubeClient, plugin.namespace)
	if err != nil {
		return err
	}

	releaseName := getHelmReleaseName(deployName)
	chartRepo, chartName, chartVersion, err := getHelmReleaseInfo(params)
	if err != nil {
		return err
	}

	helmClient, err := plugin.newClient()
	if err != nil {
		return err
	}

	chartPath, err := plugin.fetchChart(chartRepo, chartName, chartVersion)
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

	cluster := plugin.cluster
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
				plugin.namespace,
				helm.ReleaseName(releaseName),
				helm.ValueOverrides(helmParams),
				helm.InstallReuseName(true),
				helm.InstallTimeout(int64(plugin.config.Timeout)),
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
	if status.Namespace != plugin.namespace {
		return fmt.Errorf("it's not allowed to change namespace of the release %s (was %s, requested %s)", releaseName, status.Namespace, plugin.namespace)
	}

	newRelease, err := helmClient.UpdateRelease(
		releaseName,
		chartPath,
		helm.UpdateValueOverrides(helmParams),
		helm.UpgradeTimeout(int64(plugin.config.Timeout)),
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
func (plugin *Plugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := plugin.init(eventLog)
	if err != nil {
		return err
	}

	releaseName := getHelmReleaseName(deployName)

	helmClient, err := plugin.newClient()
	if err != nil {
		return err
	}

	eventLog.WithFields(event.Fields{
		"release": releaseName,
	}).Infof("Deleting Helm release '%s'", releaseName)

	_, err = helmClient.DeleteRelease(
		releaseName,
		helm.DeletePurge(true),
		helm.DeleteTimeout(int64(plugin.config.Timeout)),
	)
	return err
}

// Endpoints returns map from port type to url for all services of the current chart
func (plugin *Plugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	err := plugin.init(eventLog)
	if err != nil {
		return nil, err
	}

	kubeClient, err := plugin.kube.NewClient()
	if err != nil {
		return nil, err
	}

	client := kubeClient.CoreV1()

	releaseName := getHelmReleaseName(deployName)

	selector := labels.Set{"release": releaseName}.AsSelector().String()
	options := meta.ListOptions{LabelSelector: selector}

	endpoints := make(map[string]string)

	// Check all corresponding services
	services, err := client.Services(plugin.namespace).List(options)
	if err != nil {
		return nil, err
	}

	for _, service := range services.Items {
		// todo(slukjanov): support not only node ports
		if service.Spec.Type == "NodePort" {
			for _, port := range service.Spec.Ports {
				sURL := fmt.Sprintf("%s:%d", plugin.kube.ExternalAddress, port.NodePort)

				// todo(slukjanov): could we somehow detect real schema? I think no :(
				if util.StringContainsAny(port.Name, "https") {
					sURL = "https://" + sURL
				} else if util.StringContainsAny(port.Name, "ui", "rest", "http", "grafana") {
					sURL = "http://" + sURL
				}

				endpoints[port.Name] = sURL
			}
		}
	}

	return endpoints, nil
}
