package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"os"
	"strconv"
	"strings"
)

func (cache *clusterCache) getHTTPServicesForHelmRelease(releaseName string, chartName string, eventLog *event.Log) ([]string, error) {
	client, err := cache.newKubeClient()
	if err != nil {
		return nil, err
	}

	coreClient := client.CoreV1()

	selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector().String()
	options := meta.ListOptions{LabelSelector: selector}

	// Check all corresponding services
	services, err := coreClient.Services(cache.namespace).List(options)
	if err != nil {
		return nil, err
	}

	// Check all corresponding Istio ingresses
	ingresses, err := client.ExtensionsV1beta1().Ingresses(cache.namespace).List(options)
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

func (plugin *Plugin) getDesiredIstioRouteRulesForComponent(componentKey string, policy *lang.Policy, resolution *resolve.PolicyResolution, externalData *external.Data, eventLog *event.Log) ([]*istioRouteRule, error) {
	instance := resolution.ComponentInstanceMap[componentKey]
	serviceObj, err := policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, err
	}
	service := serviceObj.(*lang.Service)
	component := service.GetComponentsMap()[instance.Metadata.Key.ComponentName]

	calcLabels := resolution.ComponentInstanceMap[componentKey].CalculatedLabels
	clusterObj, err := policy.GetObject(lang.ClusterObject.Kind, calcLabels.Labels[lang.LabelCluster], runtime.SystemNS)
	if err != nil {
		return nil, err
	}
	cluster := clusterObj.(*lang.Cluster)

	cache, err := plugin.getClusterCache(cluster, eventLog)
	if err != nil {
		return nil, err
	}

	allows, err := strconv.ParseBool(instance.DataForPlugins[resolve.AllowIngres])
	if err != nil {
		return nil, err
	}
	// todo(slukjanov) check code type before moving forward, only helm supported
	if !allows && component != nil && component.Code != nil {
		releaseName := getHelmReleaseName(instance.GetDeployName())
		_, chartName, _, err := getHelmReleaseInfo(instance.CalculatedCodeParams)
		if err != nil {
			return nil, err
		}

		services, err := cache.getHTTPServicesForHelmRelease(releaseName, chartName, eventLog)
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

func (cache *clusterCache) getExistingIstioRouteRulesForCluster() ([]*istioRouteRule, error) {
	cmd := "get route-rules"
	rulesStr, err := cache.runIstioCmd(cmd, cache.cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get route-rules in cluster '%s' by running '%s': %s", cache.cluster.Name, cmd, err)
	}

	rules := make([]*istioRouteRule, 0)

	for _, ruleName := range strings.Split(rulesStr, "\n") {
		if ruleName == "" {
			continue
		}
		rules = append(rules, &istioRouteRule{ruleName, cache.cluster, cache})
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

	ruleFileName := util.WriteTempFile("istio-rule", []byte(content))
	defer os.Remove(ruleFileName) // nolint: errcheck

	out, err := rule.cache.runIstioCmd("create -f "+ruleFileName, rule.Cluster)
	if err != nil {
		return fmt.Errorf("failed to create istio rule in cluster '%s': %s %s", rule.Cluster.Name, out, err)
	}
	return nil
}

func (rule *istioRouteRule) destroy() error {
	out, err := rule.cache.runIstioCmd("delete route-rule "+rule.Service, rule.Cluster)
	if err != nil {
		return fmt.Errorf("failed to delete istio rule in cluster '%s': %s %s", rule.Cluster.Name, out, err)
	}
	return nil
}

func (cache *clusterCache) getIstioSvc() (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	istioSvc := cache.istioSvc
	if len(istioSvc) == 0 {
		client, err := cache.newKubeClient()
		if err != nil {
			return "", err
		}

		coreClient := client.CoreV1()

		selector := labels.Set{"app": "istio"}.AsSelector().String()
		options := meta.ListOptions{LabelSelector: selector}

		pods, err := coreClient.Pods(cache.cluster.Namespace).List(options)
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
			services, err := coreClient.Services(cache.cluster.Namespace).List(options)
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
	istioSvc, err := cache.getIstioSvc()
	if err != nil {
		return "", err
	}

	if istioSvc == "" {
		// todo(slukjanov): it's temp fix for the case when istio isn't running yet, replace it with istio polling?
		return "", nil
	}

	content := "set -e\n"
	content += "kubectl config use-context " + cluster.Name + " 1>/dev/null\n"
	content += "istioctl --configAPIService " + istioSvc + " --namespace " + cluster.Namespace + " "
	content += cmd + "\n"

	cmdFileName := util.WriteTempFile("istioctl-cmd", []byte(content))
	defer os.Remove(cmdFileName) // nolint: errcheck

	out, err := util.RunCmd("bash", cmdFileName)
	if err != nil {
		return "", err
	}

	return out, nil
}
