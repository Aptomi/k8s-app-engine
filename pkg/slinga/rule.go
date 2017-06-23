package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

// LabelsFilter is a labels filter
type LabelsFilter []string

// ServiceFilter is a service filter
type ServiceFilter struct {
	Cluster *Criteria
	Labels  *Criteria
	User    *Criteria
}

// Action is an action
type Action struct {
	Type    string
	Content string
}

// Rule is a global rule
type Rule struct {
	Enabled        bool
	Name           string
	FilterServices *ServiceFilter
	Actions        []*Action
}

// Describe return full description of the rule - conditions and actions description
func (rule *Rule) DescribeConditions() map[string][]string {
	descr := make(map[string][]string)

	if rule.FilterServices != nil {
		userFilter := rule.FilterServices.User
		if userFilter != nil {
			if len(userFilter.Accept) > 0 {
				descr["User with labels matching"] = userFilter.Accept
			}
			if len(userFilter.Reject) > 0 {
				descr["User without labels matching"] = userFilter.Reject
			}
		}
		clusterFilter := rule.FilterServices.Cluster
		if clusterFilter != nil {
			if len(clusterFilter.Accept) > 0 {
				descr["Cluster with labels matching"] = clusterFilter.Accept
			}
			if len(clusterFilter.Reject) > 0 {
				descr["Cluster without labels matching"] = clusterFilter.Reject
			}
		}
	}

	return descr
}

func (rule *Rule) DescribeActions() []string {
	descr := make([]string, 0)

	for _, action := range rule.Actions {
		if action.Type == "dependency" && action.Content == "forbid" {
			descr = append(descr, "Forbid using services")
		} else if action.Type == "ingress" && action.Content == "block" {
			descr = append(descr, "Block external access to services")
		} else {
			descr = append(descr, fmt.Sprintf("type: %s, content: %s", action.Type, action.Content))
		}
	}

	return descr
}

func (rule *Rule) MatchUser(user *User) bool {
	return rule.FilterServices != nil && rule.FilterServices.match(LabelSet{}, user, nil)
}

// UnmarshalYAML is a custom unmarshaller for Rule, which sets Enabled to True before unmarshalling
func (s *Rule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias Rule
	instance := Alias{Enabled: true}
	if err := unmarshal(&instance); err != nil {
		return err
	}
	*s = Rule(instance)
	return nil
}

// GlobalRules is a list of global rules
type GlobalRules struct {
	// action type -> []*Rule
	Rules map[string][]*Rule
}

func (globalRules *GlobalRules) allowsIngressAccess(labels LabelSet, users []*User, cluster *Cluster) bool {
	if rules, ok := globalRules.Rules["ingress"]; ok {
		for _, rule := range rules {
			// for all users of the service
			for _, user := range users {
				// TODO: this is pretty shitty that it's not a part of engine_node. so you can't even log into "rule log" (new replacement of tracing)
				if rule.FilterServices.match(labels, user, cluster) {
					for _, action := range rule.Actions {
						if action.Type == "ingress" && action.Content == "block" {
							return false
						}
					}
				}
			}
		}
	}

	return true
}

func (filter *ServiceFilter) match(labels LabelSet, user *User, cluster *Cluster) bool {
	// check if service filters for another service labels
	if filter.Labels != nil && !filter.Labels.allows(labels) {
		return false
	}

	// check if service filters for another user labels
	if filter.User != nil && !filter.User.allows(user.getLabelSet()) {
		return false
	}

	if filter.Cluster != nil && cluster != nil && !filter.Cluster.allows(cluster.getLabelSet()) {
		return false
	}

	return true
}

// NewGlobalRules creates and initializes a new empty list of global rules
func NewGlobalRules() GlobalRules {
	return GlobalRules{Rules: make(map[string][]*Rule, 0)}
}

func (globalRules *GlobalRules) addRule(rule *Rule) {
	if rule.FilterServices == nil {
		debug.WithFields(log.Fields{
			"rule": rule,
		}).Panic("Only service filters currently supported in rules")
	}
	for _, action := range rule.Actions {
		if rulesList, ok := globalRules.Rules[action.Type]; ok {
			globalRules.Rules[action.Type] = append(rulesList, rule)
		} else {
			globalRules.Rules[action.Type] = []*Rule{rule}
		}
	}
}

func (globalRules *GlobalRules) count() int {
	return countElements(globalRules.Rules)
}
