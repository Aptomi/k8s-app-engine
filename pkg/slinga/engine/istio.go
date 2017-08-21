package engine

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	log "github.com/Sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	k8slabels "k8s.io/kubernetes/pkg/labels"
	"strings"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
)

// IstioRouteRule is istio route rule
type IstioRouteRule struct {
	Service string
	Cluster *Cluster
}

// IstioRuleEnforcer enforces istio rules
type IstioRuleEnforcer struct {
	state    *ServiceUsageState
	progress progress.ProgressIndicator
}

// NewIstioRuleEnforcer creates new IstioRuleEnforcer
func NewIstioRuleEnforcer(diff *ServiceUsageStateDiff) *IstioRuleEnforcer {
	return &IstioRuleEnforcer{
		state:    diff.Next,
		progress: diff.progress,
	}
}

// Returns difference length (used for progress indicator)
func (enforcer *IstioRuleEnforcer) getDifferenceLen() int {
	result := 0

	// Get istio rules for each cluster
	result += len(enforcer.state.Policy.Clusters)

	// Call processComponent() for each component
	result += len(enforcer.state.GetResolvedData().ComponentProcessingOrder)

	// Create rules for each component
	result += len(enforcer.state.GetResolvedData().ComponentProcessingOrder)

	// Delete rules (all at once)
	result += len(enforcer.state.GetResolvedData().ComponentProcessingOrder)

	return result
}

// Apply processes global rules and applies Istio routing rules for ingresses
func (enforcer *IstioRuleEnforcer) Apply(noop bool) {
	if len(enforcer.state.GetResolvedData().ComponentProcessingOrder) == 0 || noop {
		return
	}

	existingRules := make([]*IstioRouteRule, 0)

	for _, cluster := range enforcer.state.Policy.Clusters {
		existingRules = append(existingRules, getIstioRouteRules(cluster)...)
		enforcer.progress.Advance("Istio")
	}

	// Process in the right order
	desiredRules := make(map[string][]*IstioRouteRule)
	for _, key := range enforcer.state.GetResolvedData().ComponentProcessingOrder {
		rules, err := processComponent(key, enforcer.state)
		if err != nil {
			Debug.WithFields(log.Fields{
				"key":   key,
				"error": err,
			}).Panic("Unable to process Istio Ingress for component")
		}
		desiredRules[key] = rules
		enforcer.progress.Advance("Istio")
	}

	deleteRules := make([]*IstioRouteRule, 0)
	createRules := make(map[string][]*IstioRouteRule)

	// populate createRules, to make sure we will get correct number of entries for progress indicator
	for _, key := range enforcer.state.GetResolvedData().ComponentProcessingOrder {
		createRules[key] = make([]*IstioRouteRule, 0)
	}

	for _, existingRule := range existingRules {
		found := false
		for _, desiredRulesForComponent := range desiredRules {
			for _, desiredRule := range desiredRulesForComponent {
				if existingRule.Service == desiredRule.Service && existingRule.Cluster.GetName() == desiredRule.Cluster.GetName() {
					found = true
				}
			}
		}
		if !found {
			deleteRules = append(deleteRules, existingRule)
		}
	}

	for key, desiredRulesForComponent := range desiredRules {
		for _, desiredRule := range desiredRulesForComponent {
			found := false
			for _, existingRule := range existingRules {
				if desiredRule.Service == existingRule.Service && desiredRule.Cluster.GetName() == existingRule.Cluster.GetName() {
					found = true
				}
			}
			if !found {
				createRules[key] = append(createRules[key], desiredRule)
			}
		}
	}

	// process creations by component
	changed := false
	for _, createRulesForComponent := range createRules {
		for _, rule := range createRulesForComponent {
			rule.create()
			changed = true
		}
		enforcer.progress.Advance("Istio")
	}

	// process deletions all at once
	for _, rule := range deleteRules {
		rule.delete()
		changed = true
	}
	enforcer.progress.Advance("Istio")

	if changed {
		for _, createRulesForComponent := range createRules {
			for _, rule := range createRulesForComponent {
				fmt.Println("  [+] " + rule.Service)
			}
		}
		for _, rule := range deleteRules {
			fmt.Println("  [-] " + rule.Service)
		}
	} else {
		fmt.Println("  [*] No changes")
	}
}

func processComponent(key string, usage *ServiceUsageState) ([]*IstioRouteRule, error) {
	instance := usage.GetResolvedData().ComponentInstanceMap[key]
	component := usage.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	labels := usage.GetResolvedData().ComponentInstanceMap[key].CalculatedLabels

	// todo(slukjanov): temp hack - expecting that cluster is always passed through the label "cluster"
	var cluster *Cluster
	if clusterLabel, ok := labels.Labels["cluster"]; ok {
		if cluster, ok = usage.Policy.Clusters[clusterLabel]; !ok {
			Debug.WithFields(log.Fields{
				"component": key,
				"labels":    labels.Labels,
			}).Panic("Can't find cluster for component (based on label 'cluster')")
		}
	}

	// get all users who're using service
	dependencyIds := usage.GetResolvedData().ComponentInstanceMap[key].DependencyIds
	users := make([]*User, 0)
	for dependencyID := range dependencyIds {
		// todo check if user doesn't exist
		userID := usage.Policy.Dependencies.DependenciesByID[dependencyID].UserID
		users = append(users, usage.userLoader.LoadUserByID(userID))
	}

	if !usage.Policy.Rules.AllowsIngressAccess(labels, users, cluster) && component != nil && component.Code != nil {
		codeExecutor, err := GetCodeExecutor(component.Code, key, usage.GetResolvedData().ComponentInstanceMap[key].CalculatedCodeParams, usage.Policy.Clusters)
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
	_, clientset, err := newKubeClient(exec.Cluster)
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

func getIstioRouteRules(cluster *Cluster) []*IstioRouteRule {
	cmd := "get route-rules"
	rulesStr, err := runIstioCmd(cluster, cmd)
	if err != nil {
		Debug.WithFields(log.Fields{
			"cluster": cluster.GetName(),
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

	ruleFile := WriteTempFile("istio-rule", content)

	out, err := runIstioCmd(rule.Cluster, "create -f "+ruleFile.Name())
	if err != nil {
		Debug.WithFields(log.Fields{
			"cluster": rule.Cluster.GetName(),
			"content": content,
			"out":     out,
			"error":   err,
		}).Panic("Failed to create istio rule by running bash script")
	}
}

func (rule *IstioRouteRule) delete() {
	out, err := runIstioCmd(rule.Cluster, "delete route-rule "+rule.Service)
	if err != nil {
		Debug.WithFields(log.Fields{
			"cluster": rule.Cluster.GetName(),
			"out":     out,
			"error":   err,
		}).Panic("Failed to delete istio rule by running bash script")
	}
}

func runIstioCmd(cluster *Cluster, cmd string) (string, error) {
	istioSvc := cluster.GetIstioSvc()
	if len(istioSvc) == 0 {
		_, clientset, err := newKubeClient(cluster)
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

					cluster.SetIstioSvc(istioSvc)
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
	content += "kubectl config use-context " + cluster.GetName() + " 1>/dev/null\n"
	content += "istioctl --configAPIService " + cluster.GetIstioSvc() + " --namespace " + cluster.Metadata.Namespace + " "
	content += cmd + "\n"

	cmdFile := WriteTempFile("istioctl-cmd", content)

	out, err := RunCmd("bash", cmdFile.Name())
	if err != nil {
		return "", err
	}

	return out, nil
}
