package istio

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/deployment"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/util"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
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

// IstioRuleEnforcer enforces istio rules
type IstioRuleEnforcer struct {
	next *resolve.ResolvedState
}

// NewIstioRuleEnforcer creates new IstioRuleEnforcer
func (enforcer *IstioRuleEnforcer) Init(next *resolve.ResolvedState, prev *resolve.ResolvedState) {
	enforcer.next = next
}

// Returns difference length (used for progress indicator)
func (enforcer *IstioRuleEnforcer) GetApplyProgressLength() int {
	result := 0

	// Get istio rules for each cluster
	result += len(enforcer.next.Policy.Clusters)

	// Call processComponent() for each component
	result += len(enforcer.next.State.ResolvedData.ComponentProcessingOrder)

	// Create rules for each component
	result += len(enforcer.next.State.ResolvedData.ComponentProcessingOrder)

	// Delete rules (all at once)
	result += len(enforcer.next.State.ResolvedData.ComponentProcessingOrder)

	return result
}

// Apply processes global rules and applies Istio routing rules for ingresses
func (enforcer *IstioRuleEnforcer) Apply(progress progress.ProgressIndicator) {
	if len(enforcer.next.State.ResolvedData.ComponentProcessingOrder) == 0 {
		return
	}

	existingRules := make([]*IstioRouteRule, 0)

	for _, cluster := range enforcer.next.Policy.Clusters {
		existingRules = append(existingRules, getIstioRouteRules(cluster)...)
		progress.Advance("Istio")
	}

	// Process in the right order
	desiredRules := make(map[string][]*IstioRouteRule)
	for _, key := range enforcer.next.State.ResolvedData.ComponentProcessingOrder {
		rules, err := enforcer.processComponent(key)
		if err != nil {
			Debug.WithFields(log.Fields{
				"key":   key,
				"error": err,
			}).Panic("Unable to process Istio Ingress for component")
		}
		desiredRules[key] = rules
		progress.Advance("Istio")
	}

	deleteRules := make([]*IstioRouteRule, 0)
	createRules := make(map[string][]*IstioRouteRule)

	// populate createRules, to make sure we will get correct number of entries for progress indicator
	for _, key := range enforcer.next.State.ResolvedData.ComponentProcessingOrder {
		createRules[key] = make([]*IstioRouteRule, 0)
	}

	for _, existingRule := range existingRules {
		found := false
		for _, desiredRulesForComponent := range desiredRules {
			for _, desiredRule := range desiredRulesForComponent {
				if existingRule.Service == desiredRule.Service && existingRule.Cluster.Name == desiredRule.Cluster.Name {
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
				if desiredRule.Service == existingRule.Service && desiredRule.Cluster.Name == existingRule.Cluster.Name {
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
		progress.Advance("Istio")
	}

	// process deletions all at once
	for _, rule := range deleteRules {
		rule.delete()
		changed = true
	}
	progress.Advance("Istio")

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

func (enforcer *IstioRuleEnforcer) processComponent(key string) ([]*IstioRouteRule, error) {
	usageState := enforcer.next.State

	instance := usageState.ResolvedData.ComponentInstanceMap[key]
	component := enforcer.next.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	labels := usageState.ResolvedData.ComponentInstanceMap[key].CalculatedLabels

	cluster, err := util.GetCluster(enforcer.next.Policy, labels)
	if err != nil {
		return nil, err
	}

	// get all users who're using service
	dependencyIds := usageState.ResolvedData.ComponentInstanceMap[key].DependencyIds
	users := make([]*User, 0)
	for dependencyID := range dependencyIds {
		// todo check if user doesn't exist
		userID := enforcer.next.Policy.Dependencies.DependenciesByID[dependencyID].UserID
		users = append(users, enforcer.next.UserLoader.LoadUserByID(userID))
	}

	allows, err := enforcer.next.Policy.Rules.AllowsIngressAccess(labels, users, cluster)
	if err != nil {
		return nil, err
	}
	if !allows && component != nil && component.Code != nil {
		codeExecutor, err := GetCodeExecutor(component.Code, key, usageState.ResolvedData.ComponentInstanceMap[key].CalculatedCodeParams, enforcer.next.Policy.Clusters)
		if err != nil {
			return nil, err
		}

		if helmCodeExecutor, ok := codeExecutor.(HelmCodeExecutor); ok {
			services, err := helmCodeExecutor.HttpServices()
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

func getIstioRouteRules(cluster *Cluster) []*IstioRouteRule {
	cmd := "get route-rules"
	rulesStr, err := runIstioCmd(cluster, cmd)
	if err != nil {
		Debug.WithFields(log.Fields{
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
	content += "  destination: " + rule.Service + "." + rule.Cluster.Namespace + ".svc.cluster.local\n"
	content += "  httpReqTimeout:\n"
	content += "    simpleTimeout:\n"
	content += "      timeout: 1ms\n"

	ruleFile := WriteTempFile("istio-rule", content)

	out, err := runIstioCmd(rule.Cluster, "create -f "+ruleFile.Name())
	if err != nil {
		Debug.WithFields(log.Fields{
			"cluster": rule.Cluster.Name,
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
			"cluster": rule.Cluster.Name,
			"out":     out,
			"error":   err,
		}).Panic("Failed to delete istio rule by running bash script")
	}
}

func runIstioCmd(cluster *Cluster, cmd string) (string, error) {
	istioSvc := cluster.GetIstioSvc()
	if len(istioSvc) == 0 {
		_, clientset, err := NewKubeClient(cluster)
		if err != nil {
			return "", err
		}

		coreClient := clientset.Core()

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
	content += "kubectl config use-context " + cluster.Name + " 1>/dev/null\n"
	content += "istioctl --configAPIService " + cluster.GetIstioSvc() + " --namespace " + cluster.Namespace + " "
	content += cmd + "\n"

	cmdFile := WriteTempFile("istioctl-cmd", content)

	out, err := RunCmd("bash", cmdFile.Name())
	if err != nil {
		return "", err
	}

	return out, nil
}
