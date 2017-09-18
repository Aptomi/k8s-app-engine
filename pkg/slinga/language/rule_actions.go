package language

type ChangeLabelsAction LabelOperations

type DependencyAction string

type IngressAction string

type RuleActionResult struct {
	AllowDependency bool
	AllowIngress    bool

	ChangedLabelsOnLastApply bool
	Labels                   *LabelSet
}

func NewRuleActionResult(labels *LabelSet) *RuleActionResult {
	return &RuleActionResult{
		Labels: labels,
	}
}

func (rule *Rule) ApplyActions(result *RuleActionResult) {
	if rule.Actions.Dependency != "" {
		result.AllowDependency = string(rule.Actions.Dependency) == "allow"
	}
	if rule.Actions.Ingress != "" {
		result.AllowIngress = string(rule.Actions.Ingress) == "allow"
	}
	result.ChangedLabelsOnLastApply = false
	if rule.Actions.ChangeLabels != nil {
		result.ChangedLabelsOnLastApply = result.Labels.ApplyTransform(LabelOperations(rule.Actions.ChangeLabels))
	}
}
