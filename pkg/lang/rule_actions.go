package lang

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

	Namespaces map[string]bool
}

// NewRuleActionResult creates a new RuleActionResult
func NewRuleActionResult(labels *LabelSet) *RuleActionResult {
	return &RuleActionResult{
		Labels:     labels,
		Namespaces: make(map[string]bool),
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

	for ns := range rule.Actions.Namespaces {
		result.Namespaces[ns] = true
	}
}
