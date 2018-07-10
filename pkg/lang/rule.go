package lang

import (
	"sort"

	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// TypeRule is an informational data structure with Kind and Constructor for Rule
var TypeRule = &runtime.TypeInfo{
	Kind:        "rule",
	Storable:    true,
	Versioned:   true,
	Constructor: func() runtime.Object { return &Rule{} },
}

// Rule is a generic mechanism for defining rules in Aptomi.
//
// Rules can be used to set certain labels on certain conditions as well as perform certain actions (such as rejecting
// claims, rejecting ingress traffic, etc)
type Rule struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         `validate:"required"`

	// Weight defined for the rule. All rules are sorted in the order of increasing weight and applied in that order
	Weight int `validate:"min=0"`

	// Criteria - if it gets evaluated to true during policy resolution, then rules's actions will be executed.
	// It's an optional field, so if it's nil then it is considered to be evaluated to true automatically
	Criteria *Criteria `yaml:",omitempty" validate:"omitempty"`

	// Actions define the set of actions that will be executed if Criteria gets evaluated to true
	Actions *RuleActions `validate:"required"`
}

// RuleActions is a set of actions that can be performed by a rule. All fields in this structure are optional. If a
// field is defined, then the corresponding action will be processed
type RuleActions struct {
	// ChangeLabels defines how labels should be transformed
	ChangeLabels LabelOperations `yaml:"change-labels,omitempty" validate:"omitempty,labelOperations"`

	// Claim defines whether claim should be rejected
	Claim ClaimAction `yaml:"claim,omitempty" validate:"omitempty,allowReject"`

	// Ingress defines whether ingress traffic should be rejected
	Ingress IngressAction `yaml:"ingress,omitempty" validate:"omitempty,allowReject"`
}

// Matches returns true if a rule matches
func (rule *Rule) Matches(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if rule.Criteria == nil {
		return true, nil
	}
	return rule.Criteria.allows(params, cache)
}

type ruleSorter []*Rule

func (rs ruleSorter) Len() int {
	return len(rs)
}

func (rs ruleSorter) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs ruleSorter) Less(i, j int) bool {
	return rs[i].Weight < rs[j].Weight
}

// GetRulesSortedByWeight returns all rules sorted by their weight
func GetRulesSortedByWeight(rules map[string]*Rule) []*Rule {
	result := []*Rule{}
	for _, rule := range rules {
		result = append(result, rule)
	}
	sort.Sort(ruleSorter(result))
	return result
}
