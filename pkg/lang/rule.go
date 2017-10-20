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

// Rule is a rule
type Rule struct {
	Metadata

	Weight   int `validate:"min=0"`
	Criteria *Criteria
	Actions  *RuleActions
}

// RuleActions is a set of actions performed by rule
type RuleActions struct {
	// ChangeLabels, Dependency, and Ingress fields are relevant for regular rules
	// They determine how labels should be changed, whether dependency should be allowed, and whether ingress traffic should be allowed
	ChangeLabels ChangeLabelsAction `yaml:"change-labels"`
	Dependency   DependencyAction
	Ingress      IngressAction

	// AddRole field is only relevant for ACL rules (have to keep it in this class due to the lack of generics)
	// Key in the map is role ID, while value is a set of comma-separated namespaces to which this role applies
	AddRole map[string]string `yaml:"add-role"`
}

// Matches returns if a rule matches
func (rule *Rule) Matches(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if rule.Criteria == nil {
		return true, nil
	}
	return rule.Criteria.allows(params, cache)
}

// GlobalRules is a list of global rules
type GlobalRules struct {
	// RuleMap is a map[name] -> *Rule
	RuleMap map[string]*Rule

	// Rules is an unsorted list of rules
	Rules []*Rule

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
