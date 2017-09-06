package helm_istio

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"
	"k8s.io/kubernetes/pkg/api"
	k8serrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/labels"
	"strings"
)

var helmCodeTypes = []string{"helm", "aptomi/code/kubernetes-helm"}

func (p *HelmIstioPlugin) GetSupportedCodeTypes() []string {
	return helmCodeTypes
}

func (p *HelmIstioPlugin) Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error {
	return p.createOrUpdate(cluster, deployName, params, eventLog, true)
}

func (p *HelmIstioPlugin) Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error {
	return p.createOrUpdate(cluster, deployName, params, eventLog, true)
}

func (p *HelmIstioPlugin) createOrUpdate(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog, create bool) error {
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

	chartPath, err := getValidChartPath(chartName)
	if err != nil {
		return err
	}

	helmParams, err := yaml.Marshal(params)
	if err != nil {
		return err
	}

	if create {
		exists, err := findHelmRelease(helmClient, releaseName)
		if err != nil {
			return fmt.Errorf("Error while looking for Helm release %s: %s", releaseName, err)
		}

		if exists {
			// If a release already exists, let's just go ahead and update it
			eventLog.WithFields(eventlog.Fields{}).Infof("Release '%s' already exists. Updating it", releaseName)
		}

		eventLog.WithFields(eventlog.Fields{
			"release": releaseName,
			"chart":   chartName,
			"path":    chartPath,
			"params":  string(helmParams),
		}).Infof("Installing Helm release '%s', chart '%s'", releaseName, chartName)

		_, err = helmClient.InstallRelease(chartPath, cluster.Config.Namespace, helm.ReleaseName(releaseName), helm.ValueOverrides(helmParams), helm.InstallReuseName(true))
	} else {
		eventLog.WithFields(eventlog.Fields{
			"release": releaseName,
			"chart":   chartName,
			"path":    chartPath,
			"params":  string(helmParams),
		}).Infof("Updating Helm release '%s', chart '%s'", releaseName, chartName)

		_, err = helmClient.UpdateRelease(releaseName, chartPath, helm.UpdateValueOverrides(helmParams))
	}

	return err
}

// Destroy for HelmIstioPlugin runs "helm delete" for the corresponding helm chart
func (p *HelmIstioPlugin) Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error {
	cache, err := p.getCache(cluster, eventLog)
	if err != nil {
		return err
	}

	releaseName := helmReleaseName(deployName)

	helmClient := cache.newHelmClient(cluster)

	eventLog.WithFields(eventlog.Fields{
		"release": releaseName,
	}).Infof("Deleting Helm release '%s'", releaseName)

	_, err = helmClient.DeleteRelease(releaseName, helm.DeletePurge(true))
	return err
}

// Endpoints returns map from port type to url for all services of the current chart
func (p *HelmIstioPlugin) Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) (map[string]string, error) {
	cache, err := p.getCache(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	_, client, err := cache.newKubeClient(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	coreClient := client.Core()

	releaseName := helmReleaseName(deployName)
	chartName, err := helmChartName(params)
	if err != nil {
		return nil, err
	}

	selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector()
	options := api.ListOptions{LabelSelector: selector}

	endpoints := make(map[string]string)

	// Check all corresponding services
	services, err := coreClient.Services(cluster.Config.Namespace).List(options)
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
	service, err := coreClient.Services(cluster.Config.Namespace).Get("istio-ingress")
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
	ingresses, err := client.Extensions().Ingresses(cluster.Config.Namespace).List(options)
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
