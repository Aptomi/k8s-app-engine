package lang

import (
	"strings"
)

// Allow is a special constant that is used in rule actions for allowing dependencies, ingress traffic, etc
const Allow = "allow"

// ChangeLabelsAction is a rule action to change labels
type ChangeLabelsAction LabelOperations

// DependencyAction is a rule action to allow or disallow dependency to be resolved
type DependencyAction string

// IngressAction is a rule action to to allow or disallow ingres traffic for a component
type IngressAction string

// RuleActionResult is a result of processing multiple rules on a given component
type RuleActionResult struct {
	AllowDependency bool
	AllowIngress    bool

	ChangedLabelsOnLastApply bool
	Labels                   *LabelSet

	RoleMap map[string]map[string]bool
}

// NewRuleActionResult creates a new RuleActionResult
func NewRuleActionResult(labels *LabelSet) *RuleActionResult {
	return &RuleActionResult{
		Labels:  labels,
		RoleMap: make(map[string]map[string]bool),
	}
}

// ApplyActions applies rule actions and updates result
func (rule *Rule) ApplyActions(result *RuleActionResult) {
	if rule.Actions.Dependency != "" {
		result.AllowDependency = string(rule.Actions.Dependency) == Allow
	}
	if rule.Actions.Ingress != "" {
		result.AllowIngress = string(rule.Actions.Ingress) == Allow
	}

	result.ChangedLabelsOnLastApply = false
	if rule.Actions.ChangeLabels != nil {
		result.ChangedLabelsOnLastApply = result.Labels.ApplyTransform(LabelOperations(rule.Actions.ChangeLabels))
	}

	for roleID, namespaceList := range rule.Actions.AddRole {
		role := ACLRolesMap[roleID]
		if role == nil {
			// skip non-existing roles
			continue
		}

		nsMap := result.RoleMap[roleID]
		if nsMap == nil {
			nsMap = make(map[string]bool)
			result.RoleMap[roleID] = nsMap
		}

		// mark all namespaces for the role
		namespaces := strings.Split(namespaceList, ",")
		for _, namespace := range namespaces {
			nsMap[namespace] = true
		}

		// if role covers all namespaces, mark it as well
		if role.Privileges.AllNamespaces {
			nsMap[namespaceAll] = true
		}

	}
}
