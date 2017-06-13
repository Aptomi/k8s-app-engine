package slinga

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sort"
)

type LabelsFilter []string

type ServiceFilter struct {
	Cluster *Criteria
	Labels  *Criteria
	User    *Criteria
}

type Action struct {
	Type    string
	Content string
}

type Rule struct {
	Name           string
	FilterServices *ServiceFilter
	Actions        []*Action
}

type GlobalRules struct {
	// action type -> []*Rule
	Rules map[string][]*Rule
}

func (globalRules *GlobalRules) allowsAllocation(allocation *Allocation, node *resolutionNode, cluster *Cluster) bool {
	if rules, ok := globalRules.Rules["dependency"]; ok {
		for _, rule := range rules {
			m := rule.FilterServices.match(node, cluster)
			tracing.Printf(node.depth+1, "[%t] Testing allocation '%s': (global rule '%s')", !m, allocation.Name, rule.Name)
			if m {
				for _, action := range rule.Actions {
					if action.Type == "dependency" && action.Content == "forbid" {
						return false
					}
				}
			}
		}
	}

	return true
}

func (filter *ServiceFilter) match(node *resolutionNode, cluster *Cluster) bool {
	// check if service filters for another service labels
	if filter.Labels != nil && !filter.Labels.allows(node.labels) {
		return false
	}

	// check if service filters for another user labels
	if filter.User != nil && !filter.User.allows(node.user.getLabelSet()) {
		return false
	}

	if filter.Cluster != nil && cluster != nil && !filter.Cluster.allows(cluster.getLabelSet()) {
		return false
	}

	return true
}

func NewGlobalRules() GlobalRules {
	return GlobalRules{Rules: make(map[string][]*Rule, 0)}
}

func LoadRulesFromDir(dir string) GlobalRules {
	files, _ := zglob.Glob(dir + "/**/rules.*.yaml")
	sort.Strings(files)
	r := NewGlobalRules()
	for _, f := range files {
		rules := LoadRulesFromFile(f)
		r.insertRules(rules...)
	}
	return r
}

func (rules *GlobalRules) insertRules(appendRules ...*Rule) {
	for _, rule := range appendRules {
		if rule.FilterServices == nil {
			debug.WithFields(log.Fields{
				"rule": rule,
			}).Fatal("Only service filters currently supported in rules")
		}
		for _, action := range rule.Actions {
			if rulesList, ok := rules.Rules[action.Type]; ok {
				rules.Rules[action.Type] = append(rulesList, rule)
			} else {
				rules.Rules[action.Type] = []*Rule{rule}
			}
		}
	}
}

func LoadRulesFromFile(fileName string) []*Rule {
	debug.WithFields(log.Fields{
		"file": fileName,
	}).Debug("Loading rules")

	dat, e := ioutil.ReadFile(fileName)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	t := []*Rule{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal rules")
	}
	return t
}
