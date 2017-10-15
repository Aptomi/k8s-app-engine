package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"gopkg.in/yaml.v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/helm/pkg/helm"
	"strings"
)

var helmCodeTypes = []string{"helm", "aptomi/code/kubernetes-helm"}

// GetSupportedCodeTypes returns all code types for which this plugin is registered to
func (p *Plugin) GetSupportedCodeTypes() []string {
	return helmCodeTypes
}

// Create creates component instance in the cloud by deploying a helm chart
func (p *Plugin) Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	return p.createOrUpdate(cluster, deployName, params, eventLog, true)
}

// Update updates component instance in the cloud by updating parameters of a helm chart
func (p *Plugin) Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	return p.createOrUpdate(cluster, deployName, params, eventLog, true)
}

func (p *Plugin) createOrUpdate(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log, create bool) error {
	cache, err := p.getCache(cluster, eventLog)
	if err != nil {
		return err
	}

	releaseName := helmReleaseName(deployName)
	chartName, err := helmChartName(params)
	if err != nil {
		return err
	}

	helmClient := cache.newHelmClient(cluster)

	chartPath, err := p.getValidChartPath(chartName)
	if err != nil {
		return err
	}

	helmParams, err := yaml.Marshal(params)
	if err != nil {
		return err
	}

	if create {
		exists, errRelease := findHelmRelease(helmClient, releaseName)
		if errRelease != nil {
			return fmt.Errorf("Error while looking for Helm release %s: %s", releaseName, errRelease)
		}

		if exists {
			// If a release already exists, let's just go ahead and update it
			eventLog.WithFields(event.Fields{}).Infof("Release '%s' already exists. Updating it", releaseName)
		} else {
			eventLog.WithFields(event.Fields{
				"release": releaseName,
				"chart":   chartName,
				"path":    chartPath,
				"params":  string(helmParams),
			}).Infof("Installing Helm release '%s', chart '%s', cluster: '%s'", releaseName, chartName, cluster.Name)

			_, err = helmClient.InstallRelease(chartPath, cluster.Config.Namespace, helm.ReleaseName(releaseName), helm.ValueOverrides(helmParams), helm.InstallReuseName(true))

			return err
		}
	}

	eventLog.WithFields(event.Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(helmParams),
	}).Infof("Updating Helm release '%s', chart '%s', cluster: '%s'", releaseName, chartName, cluster.Name)

	_, err = helmClient.UpdateRelease(releaseName, chartPath, helm.UpdateValueOverrides(helmParams))

	return err
}

// Destroy for Plugin runs "helm delete" for the corresponding helm chart
func (p *Plugin) Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	cache, err := p.getCache(cluster, eventLog)
	if err != nil {
		return err
	}

	releaseName := helmReleaseName(deployName)

	helmClient := cache.newHelmClient(cluster)

	eventLog.WithFields(event.Fields{
		"release": releaseName,
	}).Infof("Deleting Helm release '%s'", releaseName)

	_, err = helmClient.DeleteRelease(releaseName, helm.DeletePurge(true))
	return err
}

// Endpoints returns map from port type to url for all services of the current chart
func (p *Plugin) Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	cache, err := p.getCache(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	_, kubeClient, err := cache.newKubeClient(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	client := kubeClient.CoreV1()

	releaseName := helmReleaseName(deployName)
	chartName, err := helmChartName(params)
	if err != nil {
		return nil, err
	}

	selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector().String()
	options := meta.ListOptions{LabelSelector: selector}

	endpoints := make(map[string]string)

	// Check all corresponding services
	services, err := client.Services(cluster.Config.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	kubeHost, err := cache.getKubeExternalAddress(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	for _, service := range services.Items {
		if service.Spec.Type == "NodePort" {
			for _, port := range service.Spec.Ports {
				sURL := fmt.Sprintf("%s:%d", kubeHost, port.NodePort)

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

	// Find Istio Ingress service (how ingress itself exposed)
	service, err := client.Services(cluster.Config.Namespace).Get("istio-ingress", meta.GetOptions{})
	if err != nil {
		// return if there is no Istio deployed
		if k8serrors.IsNotFound(err) {
			return endpoints, nil
		}
		return nil, err
	}

	istioIngress := "<unresolved>"
	if service.Spec.Type == "NodePort" {
		for _, port := range service.Spec.Ports {
			if port.Name == "http" {
				istioIngress = fmt.Sprintf("%s:%d", kubeHost, port.NodePort)
			}
		}
	}

	// Check all corresponding istio ingresses
	ingresses, err := kubeClient.Extensions().Ingresses(cluster.Config.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	// todo(slukjanov): support more then one ingress / rule / path
	for _, ingress := range ingresses.Items {
		if class, ok := ingress.Annotations["kubernetes.io/ingress.class"]; !ok || class != "istio" {
			continue
		}
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				pathStr := strings.Trim(path.Path, ".*")

				if rule.Host == "" {
					endpoints["ingress"] = "http://" + istioIngress + pathStr
				} else {
					endpoints["ingress"] = "http://" + rule.Host + pathStr
				}
			}
		}
	}

	return endpoints, nil
}
