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
}

type Action struct {
	Type    string
	Content string
}

type Rule struct {
	Name           string
	FilterServices *ServiceFilter
	Actions        []*Action

	// This field is populated when dependency gets resolved
	//ResolvesTo string
}

type GlobalRules struct {
	// action type -> []*Rule
	Rules map[string][]*Rule
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
