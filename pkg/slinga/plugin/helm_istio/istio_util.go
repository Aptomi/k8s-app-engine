package helm_istio

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
	k8slabels "k8s.io/kubernetes/pkg/labels"
	"os"
	"strconv"
	"strings"
)

func (cache *clusterCache) getHttpServicesForHelmRelease(cluster *lang.Cluster, releaseName string, chartName string, eventLog *eventlog.EventLog) ([]string, error) {
	_, client, err := cache.newKubeClient(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	coreClient := client.Core()

	selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector()
	options := api.ListOptions{LabelSelector: selector}

	// Check all corresponding services
	services, err := coreClient.Services(cluster.Config.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	// Check all corresponding Istio ingresses
	ingresses, err := client.Extensions().Ingresses(cluster.Config.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	if len(ingresses.Items) > 0 {
		result := make([]string, 0)
		for _, service := range services.Items {
			result = append(result, service.Name)
		}

		return result, nil
	}

	return nil, nil
}

func (p *HelmIstioPlugin) getDesiredIstioRouteRulesForComponent(componentKey string, policy *lang.Policy, resolution *resolve.PolicyResolution, externalData *external.Data, eventLog *eventlog.EventLog) ([]*istioRouteRule, error) {
	instance := resolution.ComponentInstanceMap[componentKey]
	component := policy.Services[instance.Metadata.Key.ServiceName].GetComponentsMap()[instance.Metadata.Key.ComponentName]

	calcLabels := resolution.ComponentInstanceMap[componentKey].CalculatedLabels
	cluster, err := policy.GetClusterByLabels(calcLabels)
	if err != nil {
		return nil, err
	}

	cache, err := p.getCache(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	// get all users who're using service
	dependencyIds := resolution.ComponentInstanceMap[componentKey].DependencyIds
	users := make([]*lang.User, 0)
	for dependencyID := range dependencyIds {
		// todo check if user doesn't exist
		userID := policy.Dependencies.DependenciesByID[dependencyID].UserID
		users = append(users, externalData.UserLoader.LoadUserByID(userID))
	}

	allows, err := strconv.ParseBool(instance.DataForPlugins[resolve.ALLOW_INGRESS])
	if err != nil {
		return nil, err
	}
	// todo(slukjanov) check code type before moving forward, only helm supported
	if !allows && component != nil && component.Code != nil {
		releaseName := helmReleaseName(componentKey)
		chartName, err := helmChartName(instance.CalculatedCodeParams)
		if err != nil {
			return nil, err
		}

		services, err := cache.getHttpServicesForHelmRelease(cluster, releaseName, chartName, eventLog)
		if err != nil {
			return nil, err
		}

		rules := make([]*istioRouteRule, 0)

		for _, service := range services {
			rules = append(rules, &istioRouteRule{service, cluster, cache})
		}

		return rules, nil
	}

	return nil, nil
}

func (cache *clusterCache) getExistingIstioRouteRulesForCluster(cluster *lang.Cluster) ([]*istioRouteRule, error) {
	cmd := "get route-rules"
	rulesStr, err := cache.runIstioCmd(cmd, cluster)
	if err != nil {
		return nil, fmt.Errorf("Failed to get route-rules in cluster '%s' by running '%s': %s", cluster.Name, cmd, err.Error())
	}

	rules := make([]*istioRouteRule, 0)

	for _, ruleName := range strings.Split(rulesStr, "\n") {
		if ruleName == "" {
			continue
		}
		rules = append(rules, &istioRouteRule{ruleName, cluster, cache})
	}

	return rules, nil
}

type istioRouteRule struct {
	Service string
	Cluster *lang.Cluster
	cache   *clusterCache
}

func (rule *istioRouteRule) create() error {
	content := "type: route-rule\n"
	content += "name: " + rule.Service + "\n"
	content += "spec:\n"
	content += "  destination: " + rule.Service + "." + rule.Cluster.Namespace + ".svc.cluster.local\n"
	content += "  httpReqTimeout:\n"
	content += "    simpleTimeout:\n"
	content += "      timeout: 1ms\n"

	ruleFileName := util.WriteTempFile("istio-rule", content)
	defer os.Remove(ruleFileName)

	out, err := rule.cache.runIstioCmd("create -f "+ruleFileName, rule.Cluster)
	if err != nil {
		return fmt.Errorf("Failed to create istio rule in cluster '%s': %s %s", rule.Cluster.Name, out, err)
	}
	return nil
}

func (rule *istioRouteRule) destroy() error {
	out, err := rule.cache.runIstioCmd("delete route-rule "+rule.Service, rule.Cluster)
	if err != nil {
		return fmt.Errorf("Failed to delete istio rule in cluster '%s': %s %s", rule.Cluster.Name, out, err)
	}
	return nil
}

func (cache *clusterCache) getIstioSvc(cluster *lang.Cluster, eventLog *eventlog.EventLog) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	istioSvc := cache.istioSvc
	if len(istioSvc) == 0 {
		_, client, err := cache.newKubeClient(cluster, eventLog)
		if err != nil {
			return "", err
		}

		coreClient := client.Core()

		selector := k8slabels.Set{"app": "istio"}.AsSelector()
		options := api.ListOptions{LabelSelector: selector}

		pods, err := coreClient.Pods(cluster.Namespace).List(options)
		if err != nil {
			return "", err
		}

		running := false
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "istio-pilot") {
				if pod.Status.Phase == "Running" {
					contReady := true
					for _, cont := range pod.Status.ContainerStatuses {
						if !cont.Ready {
							contReady = false
						}
					}
					if contReady {
						running = true
						break
					}
				}
			}
		}

		if running {
			services, err := coreClient.Services(cluster.Namespace).List(options)
			if err != nil {
				return "", err
			}

			for _, service := range services.Items {
				if strings.Contains(service.Name, "istio-pilot") {
					istioSvc = service.Name

					for _, port := range service.Spec.Ports {
						if port.Name == "http-apiserver" {
							istioSvc = fmt.Sprintf("%s:%d", istioSvc, port.Port)
							break
						}
					}

					cache.istioSvc = istioSvc
					break
				}
			}
		}
	}

	return cache.istioSvc, nil
}

func (cache *clusterCache) runIstioCmd(cmd string, cluster *lang.Cluster) (string, error) {
	istioSvc := cache.istioSvc
	if istioSvc == "" {
		// todo(slukjanov): it's temp fix for the case when istio isn't running yet, replace it with istio polling?
		return "", nil
	}

	content := "set -e\n"
	content += "kubectl config use-context " + cluster.Name + " 1>/dev/null\n"
	content += "istioctl --configAPIService " + istioSvc + " --namespace " + cluster.Namespace + " "
	content += cmd + "\n"

	cmdFileName := util.WriteTempFile("istioctl-cmd", content)
	defer os.Remove(cmdFileName)

	out, err := util.RunCmd("bash", cmdFileName)
	if err != nil {
		return "", err
	}

	return out, nil
}
