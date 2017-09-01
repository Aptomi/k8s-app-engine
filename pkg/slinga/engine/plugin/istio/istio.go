package istio

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/base"
	. "github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/deployment"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/util"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"k8s.io/kubernetes/pkg/api"
	k8slabels "k8s.io/kubernetes/pkg/labels"
	"os"
	"strings"
)

// IstioRouteRule is istio route rule
type IstioRouteRule struct {
	Service string
	Cluster *Cluster
}

// RuleEnforcerPlugin enforces istio rules
type RuleEnforcerPlugin struct {
	*base.BasePlugin
}

// Returns difference length (used for progress indicator)
func (enforcer *RuleEnforcerPlugin) GetCustomApplyProgressLength() int {
	result := 0

	// Get istio rules for each cluster
	result += len(enforcer.Desired.Policy.Clusters)

	// Call processComponent() for each component
	result += len(enforcer.Desired.Resolution.Resolved.ComponentProcessingOrder)

	// Create rules for each component
	result += len(enforcer.Desired.Resolution.Resolved.ComponentProcessingOrder)

	// Delete rules (all at once)
	result += len(enforcer.Desired.Resolution.Resolved.ComponentProcessingOrder)

	return result
}

// Apply processes global rules and applies Istio routing rules for ingresses
func (enforcer *RuleEnforcerPlugin) OnApplyCustom(progress progress.ProgressIndicator) error {
	if len(enforcer.Desired.Resolution.Resolved.ComponentProcessingOrder) == 0 {
		return nil
	}

	enforcer.EventLog.WithFields(Fields{}).Info("Figuring out which Istio rules have to be added/deleted")

	existingRules := make([]*IstioRouteRule, 0)

	for _, cluster := range enforcer.Desired.Policy.Clusters {
		rules, err := getIstioRouteRules(cluster)
		if err != nil {
			return err
		}
		existingRules = append(existingRules, rules...)
		progress.Advance("Istio")
	}

	// Process in the right order
	desiredRules := make(map[string][]*IstioRouteRule)
	for _, key := range enforcer.Desired.Resolution.Resolved.ComponentProcessingOrder {
		rules, err := enforcer.processComponent(key)
		if err != nil {
			return fmt.Errorf("Error while processing Istio Ingress for component '%s': %s", key, err.Error())
		}
		desiredRules[key] = rules
		progress.Advance("Istio")
	}

	deleteRules := make([]*IstioRouteRule, 0)
	createRules := make(map[string][]*IstioRouteRule)

	// populate createRules, to make sure we will get correct number of entries for progress indicator
	for _, key := range enforcer.Desired.Resolution.Resolved.ComponentProcessingOrder {
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
			enforcer.EventLog.WithFields(Fields{}).Infof("Creating Istio rule: %s (%s)", rule.Service, rule.Cluster.Name)
			err := rule.create()
			if err != nil {
				return err
			}
			changed = true
		}
		progress.Advance("Istio")
	}

	// process deletions all at once
	for _, rule := range deleteRules {
		enforcer.EventLog.WithFields(Fields{}).Infof("Deleting Istio rule: %s (%s)", rule.Service, rule.Cluster.Name)
		err := rule.delete()
		if err != nil {
			return err
		}
		changed = true
	}
	progress.Advance("Istio")

	if changed {
		enforcer.EventLog.WithFields(Fields{}).Infof("Successfully processed Istio rules")
	} else {
		enforcer.EventLog.WithFields(Fields{}).Infof("No changes in Istio rules")
	}

	return nil
}

func (enforcer *RuleEnforcerPlugin) processComponent(key string) ([]*IstioRouteRule, error) {
	resolution := enforcer.Desired.Resolution

	instance := resolution.Resolved.ComponentInstanceMap[key]
	component := enforcer.Desired.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	labels := resolution.Resolved.ComponentInstanceMap[key].CalculatedLabels

	cluster, err := util.GetCluster(enforcer.Desired.Policy, labels)
	if err != nil {
		return nil, err
	}

	// get all users who're using service
	dependencyIds := resolution.Resolved.ComponentInstanceMap[key].DependencyIds
	users := make([]*User, 0)
	for dependencyID := range dependencyIds {
		// todo check if user doesn't exist
		userID := enforcer.Desired.Policy.Dependencies.DependenciesByID[dependencyID].UserID
		users = append(users, enforcer.UserLoader.LoadUserByID(userID))
	}

	allows, err := enforcer.Desired.Policy.Rules.AllowsIngressAccess(labels, users, cluster)
	if err != nil {
		return nil, err
	}
	if !allows && component != nil && component.Code != nil {
		codeExecutor, err := GetCodeExecutor(component.Code, key, resolution.Resolved.ComponentInstanceMap[key].CalculatedCodeParams, enforcer.Desired.Policy.Clusters, enforcer.EventLog)
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

func getIstioRouteRules(cluster *Cluster) ([]*IstioRouteRule, error) {
	cmd := "get route-rules"
	rulesStr, err := runIstioCmd(cluster, cmd)
	if err != nil {
		return nil, fmt.Errorf("Failed to get route-rules in cluster '%s' by running '%s': %s", cluster.Name, cmd, err.Error())
	}

	rules := make([]*IstioRouteRule, 0)

	for _, ruleName := range strings.Split(rulesStr, "\n") {
		if ruleName == "" {
			continue
		}
		rules = append(rules, &IstioRouteRule{ruleName, cluster})
	}

	return rules, nil
}

func (rule *IstioRouteRule) create() error {
	content := "type: route-rule\n"
	content += "name: " + rule.Service + "\n"
	content += "spec:\n"
	content += "  destination: " + rule.Service + "." + rule.Cluster.Namespace + ".svc.cluster.local\n"
	content += "  httpReqTimeout:\n"
	content += "    simpleTimeout:\n"
	content += "      timeout: 1ms\n"

	ruleFileName := WriteTempFile("istio-rule", content)
	defer os.Remove(ruleFileName)

	out, err := runIstioCmd(rule.Cluster, "create -f "+ruleFileName)
	if err != nil {
		return fmt.Errorf("Failed to create istio rule in cluster '%s': %s %s", rule.Cluster.Name, out, err.Error())
	}
	return nil
}

func (rule *IstioRouteRule) delete() error {
	out, err := runIstioCmd(rule.Cluster, "delete route-rule "+rule.Service)
	if err != nil {
		return fmt.Errorf("Failed to delete istio rule in cluster '%s': %s %s", rule.Cluster.Name, out, err.Error())
	}
	return nil
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

	cmdFileName := WriteTempFile("istioctl-cmd", content)
	defer os.Remove(cmdFileName)

	out, err := RunCmd("bash", cmdFileName)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (enforcer *RuleEnforcerPlugin) OnApplyComponentInstanceCreate(key string) error {
	return nil
}

func (enforcer *RuleEnforcerPlugin) OnApplyComponentInstanceUpdate(key string) error {
	return nil
}

func (enforcer *RuleEnforcerPlugin) OnApplyComponentInstanceDelete(key string) error {
	return nil
}
