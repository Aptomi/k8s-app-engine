package lang

// Reject is a special constant that is used in rule actions for rejecting claims, ingress traffic, etc
const Reject = "reject"

// ClaimAction is a rule action to allow or disallow claim to be resolved
type ClaimAction string

// IngressAction is a rule action to to allow or disallow ingres traffic for a component
type IngressAction string

// RuleActionResult is a result of processing multiple rules on a given component
type RuleActionResult struct {
	RejectClaim   bool
	RejectIngress bool

	ChangedLabelsOnLastApply bool
	Labels                   *LabelSet
}

// NewRuleActionResult creates a new RuleActionResult
func NewRuleActionResult(labels *LabelSet) *RuleActionResult {
	return &RuleActionResult{
		Labels: labels,
	}
}

// ApplyActions applies rule actions and updates result
func (rule *Rule) ApplyActions(result *RuleActionResult) {
	result.RejectClaim = string(rule.Actions.Claim) == Reject
	result.RejectIngress = string(rule.Actions.Ingress) == Reject

	result.ChangedLabelsOnLastApply = false
	if rule.Actions.ChangeLabels != nil {
		result.ChangedLabelsOnLastApply = result.Labels.ApplyTransform(rule.Actions.ChangeLabels)
	}
}
