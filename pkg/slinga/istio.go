package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	k8slabels "k8s.io/kubernetes/pkg/labels"
	"strings"
)

// IstioRouteRule is istio route rule
type IstioRouteRule struct {
	Service string
	Cluster *Cluster
}

// ProcessIstioIngress processes global rules and applies Istio routing rules for ingresses
func (usage *ServiceUsageState) ProcessIstioIngress(noop bool) {
	if len(usage.GetResolvedUsage().ComponentProcessingOrder) == 0 || noop {
		return
	}

	fmt.Println("[Route Rules (Istio)]")

	progress := NewProgress()
	progressBar := AddProgressBar(progress, len(usage.GetResolvedUsage().ComponentProcessingOrder)+len(usage.Policy.Clusters))

	existingRules := make([]*IstioRouteRule, 0)

	for _, cluster := range usage.Policy.Clusters {
		existingRules = append(existingRules, cluster.getIstioRouteRules()...)

		progressBar.Incr()
	}

	desiredRules := make([]*IstioRouteRule, 0)

	// Process in the right order
	for _, key := range usage.GetResolvedUsage().ComponentProcessingOrder {
		rules, err := processComponent(key, usage)
		if err != nil {
			debug.WithFields(log.Fields{
				"key":   key,
				"error": err,
			}).Panic("Unable to process Istio Ingress for component")
		}
		desiredRules = append(desiredRules, rules...)
		progressBar.Incr()
	}

	deleteRules := make([]*IstioRouteRule, 0)
	createRules := make([]*IstioRouteRule, 0)

	for _, existingRule := range existingRules {
		found := false
		for _, desiredRule := range desiredRules {
			if existingRule.Service == desiredRule.Service && existingRule.Cluster.Name == desiredRule.Cluster.Name {
				found = true
			}
		}
		if !found {
			deleteRules = append(deleteRules, existingRule)
		}
	}

	for _, desiredRule := range desiredRules {
		found := false
		for _, existingRule := range existingRules {
			if desiredRule.Service == existingRule.Service && desiredRule.Cluster.Name == existingRule.Cluster.Name {
				found = true
			}
		}
		if !found {
			createRules = append(createRules, desiredRule)
		}
	}

	tbdLen := len(createRules) + len(deleteRules)

	if tbdLen > 0 {
		progressBar = AddProgressBar(progress, tbdLen)

		for _, rule := range createRules {
			rule.create()
			progressBar.Incr()
		}

		for _, rule := range deleteRules {
			rule.delete()
			progressBar.Incr()
		}
	}

	progress.Stop()

	if tbdLen > 0 {
		for _, rule := range deleteRules {
			fmt.Println("  [-] " + rule.Service)
		}
		for _, rule := range createRules {
			fmt.Println("  [+] " + rule.Service)
		}
	} else {
		fmt.Println("  [*] No changes")
	}
}

func processComponent(key string, usage *ServiceUsageState) ([]*IstioRouteRule, error) {
	serviceName, _, _, componentName := ParseServiceUsageKey(key)
	component := usage.Policy.Services[serviceName].getComponentsMap()[componentName]

	labels := usage.GetResolvedUsage().ComponentInstanceMap[key].CalculatedLabels

	// todo(slukjanov): temp hack - expecting that cluster is always passed through the label "cluster"
	var cluster *Cluster
	if clusterLabel, ok := labels.Labels["cluster"]; ok {
		if cluster, ok = usage.Policy.Clusters[clusterLabel]; !ok {
			debug.WithFields(log.Fields{
				"component": key,
				"labels":    labels.Labels,
			}).Panic("Can't find cluster for component (based on label 'cluster')")
		}
	}

	// get all users who're using service
	dependencyIds := usage.GetResolvedUsage().ComponentInstanceMap[key].DependencyIds
	users := make([]*User, 0)
	for _, dependencyID := range dependencyIds {
		// todo check if user doesn't exists
		userID := usage.Dependencies.DependenciesByID[dependencyID].UserID
		users = append(users, usage.users.Users[userID])
	}

	if !usage.Policy.Rules.allowsIngressAccess(labels, users, cluster) && component != nil && component.Code != nil {
		codeExecutor, err := component.Code.GetCodeExecutor(key, usage.GetResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, usage.Policy.Clusters)
		if err != nil {
			return nil, err
		}

		if helmCodeExecutor, ok := codeExecutor.(HelmCodeExecutor); ok {
			services, err := helmCodeExecutor.httpServices()
			if err != nil {
				return nil, err
			}

			rules := make([]*IstioRouteRule, 0)

			for _, service := range services {
				rules = append(rules, &IstioRouteRule{service, cluster})
			}

			return rules, nil
		}
	}

	return nil, nil
}

// httpServices returns list of services for the current chart
func (exec HelmCodeExecutor) httpServices() ([]string, error) {
	_, clientset, err := exec.Cluster.newKubeClient()
	if err != nil {
		return nil, err
	}

	coreClient := clientset.Core()

	releaseName := releaseName(exec.Key)
	chartName := exec.chartName()

	selector := k8slabels.Set{"release": releaseName, "chart": chartName}.AsSelector()
	options := api.ListOptions{LabelSelector: selector}

	// Check all corresponding services
	services, err := coreClient.Services(exec.Cluster.Metadata.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	// Check all corresponding Istio ingresses
	ingresses, err := clientset.Extensions().Ingresses(exec.Cluster.Metadata.Namespace).List(options)
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

func (cluster *Cluster) getIstioRouteRules() []*IstioRouteRule {
	cmd := "get route-rules"
	rulesStr, err := cluster.runIstioCmd(cmd)
	if err != nil {
		debug.WithFields(log.Fields{
			"cluster": cluster.Name,
			"cmd":     cmd,
			"error":   err,
		}).Panic("Failed to get route-rules by running bash cmd")
	}

	rules := make([]*IstioRouteRule, 0)

	for _, ruleName := range strings.Split(rulesStr, "\n") {
		if ruleName == "" {
			continue
		}
		rules = append(rules, &IstioRouteRule{ruleName, cluster})
	}

	return rules
}

func (rule *IstioRouteRule) create() {
	content := "type: route-rule\n"
	content += "name: " + rule.Service + "\n"
	content += "spec:\n"
	content += "  destination: " + rule.Service + "." + rule.Cluster.Metadata.Namespace + ".svc.cluster.local\n"
	content += "  httpReqTimeout:\n"
	content += "    simpleTimeout:\n"
	content += "      timeout: 1ms\n"

	ruleFile := writeTempFile("istio-rule", content)

	out, err := rule.Cluster.runIstioCmd("create -f " + ruleFile.Name())
	if err != nil {
		debug.WithFields(log.Fields{
			"cluster": rule.Cluster.Name,
			"content": content,
			"out":     out,
			"error":   err,
		}).Panic("Failed to create istio rule by running bash script")
	}
}

func (rule *IstioRouteRule) delete() {
	out, err := rule.Cluster.runIstioCmd("delete route-rule " + rule.Service)
	if err != nil {
		debug.WithFields(log.Fields{
			"cluster": rule.Cluster.Name,
			"out":     out,
			"error":   err,
		}).Panic("Failed to delete istio rule by running bash script")
	}
}

func (cluster *Cluster) runIstioCmd(cmd string) (string, error) {
	istioSvc := cluster.Metadata.istioSvc
	if istioSvc == "" {
		_, clientset, err := cluster.newKubeClient()
		if err != nil {
			return "", err
		}

		coreClient := clientset.Core()

		selector := k8slabels.Set{"app": "istio"}.AsSelector()
		options := api.ListOptions{LabelSelector: selector}

		pods, err := coreClient.Pods(cluster.Metadata.Namespace).List(options)
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
			services, err := coreClient.Services(cluster.Metadata.Namespace).List(options)
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

					cluster.Metadata.istioSvc = istioSvc
					break
				}
			}
		}
	}

	if istioSvc == "" {
		// todo(slukjanov): it's temp fix for the case when istio isn't running yet
		return "", nil
	}

	content := "set -e\n"
	content += "kubectl config use-context " + cluster.Name + " 1>/dev/null\n"
	content += "istioctl --configAPIService " + cluster.Metadata.istioSvc + " --namespace " + cluster.Metadata.Namespace + " "
	content += cmd + "\n"

	cmdFile := writeTempFile("istioctl-cmd", content)

	out, err := runCmd("bash", cmdFile.Name())
	if err != nil {
		return "", err
	}

	return out, nil
}
