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
	Labels *Criteria
}

type Action map[string]string

type Rule struct {
	Name string
	FilterServices *ServiceFilter
	Action []*Action

	// This field is populated when dependency gets resolved
	//ResolvesTo string
}

type GlobalRules struct {
	ServiceRules []*Rule
}

func NewGlobalRules() GlobalRules {
	return GlobalRules{ServiceRules: make([]*Rule, 0)}
}

func LoadRulesFromDir(dir string) GlobalRules {
	files, _ := zglob.Glob(dir + "/**/rules.*.yaml")
	sort.Strings(files)
	r := NewGlobalRules()
	for _, f := range files {
		rules := LoadRulesFromFile(f)
		r.ServiceRules = append(r.ServiceRules, rules.ServiceRules...)
	}
	return r
}

func LoadRulesFromFile(fileName string) GlobalRules {
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
	r := NewGlobalRules()
	for _, rule := range t {
		if rule.FilterServices != nil {
			r.ServiceRules = append(r.ServiceRules, rule)
		} else {
			debug.WithFields(log.Fields{
				"file":  fileName,
				"error": e,
			}).Fatal("Only service filters currently supported in rules")
		}
	}
	return r
}
