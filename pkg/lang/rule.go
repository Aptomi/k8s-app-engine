package lang

import (
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/object"
	"sort"
	"sync"
)

// RuleObject is an informational data structure with Kind and Constructor for Rule
var RuleObject = &object.Info{
	Kind:        "rule",
	Versioned:   true,
	Constructor: func() object.Base { return &Rule{} },
}

// Rule is a generic mechanism for defining rules in Aptomi.
//
// Rules can be used to set certain labels on certain conditions as well as perform certain actions (such as rejecting
// dependencies, rejecting ingress traffic, etc)
//
// ACLRule is inherited from Rule, so the same mechanism is used for processing ACLs in Aptomi.
type Rule struct {
	Metadata `validate:"required"`

	// Weight defined for the rule. All rules are sorted in the order of increasing weight and applied in that order
	Weight int `validate:"min=0"`

	// Criteria - if it gets evaluated to true during policy resolution, then rules's actions will be executed.
	// It's an optional field, so if it's nil then it is considered to be evaluated to true automatically
	Criteria *Criteria `validate:"omitempty"`

	// Actions define the set of actions that will be executed if Criteria gets evaluated to true
	Actions *RuleActions `validate:"required"`
}

// RuleActions is a set of actions that can be performed by a rule. All fields in this structure are optional. If a
// field is defined, then the corresponding action will be processed
type RuleActions struct {
	// ChangeLabels defines how labels should be transformed
	ChangeLabels LabelOperations `yaml:"change-labels" validate:"omitempty,labelOperations"`

	// Dependency defines whether dependency should be rejected
	Dependency DependencyAction `validate:"omitempty,allowReject"`

	// Ingress defines whether ingress traffic should be rejected
	Ingress IngressAction `validate:"omitempty,allowReject"`

	// AddRole field is only relevant for ACL rules (have to keep it in this class due to the lack of generics).
	// Key in the map is role ID, while value is a set of comma-separated namespaces to which this role applies
	AddRole map[string]string `yaml:"add-role" validate:"omitempty,addRoleNS"`
}

// Matches returns true if a rule matches
func (rule *Rule) Matches(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if rule.Criteria == nil {
		return true, nil
	}
	return rule.Criteria.allows(params, cache)
}

// GlobalRules contains a map of global rules by name, as well as the list of sorted rules
type GlobalRules struct {
	// RuleMap is a map[name] -> *Rule
	RuleMap map[string]*Rule `validate:"dive"`

	// Rules is an unsorted list of rules
	Rules []*Rule `validate:"-"`

	once        sync.Once
	rulesSorted []*Rule // lazily initialized value
}

// NewGlobalRules creates and initializes a new empty list of global rules
func NewGlobalRules() *GlobalRules {
	return &GlobalRules{
		RuleMap: make(map[string]*Rule),
	}
}

func (globalRules *GlobalRules) addRule(rule ...*Rule) {
	for _, r := range rule {
		globalRules.RuleMap[r.GetName()] = r
	}

	globalRules.Rules = append(globalRules.Rules, rule...)
}

// GetRulesSortedByWeight returns all rules sorted by weight
func (globalRules *GlobalRules) GetRulesSortedByWeight() []*Rule {
	globalRules.once.Do(func() {
		globalRules.rulesSorted = append(globalRules.rulesSorted, globalRules.Rules...)
		sort.Sort(globalRules)
	})
	return globalRules.rulesSorted
}

func (globalRules *GlobalRules) Len() int {
	return len(globalRules.rulesSorted)
}

func (globalRules *GlobalRules) Less(i, j int) bool {
	return globalRules.rulesSorted[i].Weight < globalRules.rulesSorted[j].Weight
}

func (globalRules *GlobalRules) Swap(i, j int) {
	globalRules.rulesSorted[i], globalRules.rulesSorted[j] = globalRules.rulesSorted[j], globalRules.rulesSorted[i]
}
