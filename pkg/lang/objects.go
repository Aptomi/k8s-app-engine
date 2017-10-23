package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

var (
	// Objects is a list of object.Info for all lang objects
	Objects = []*object.Info{
		ServiceObject,
		ContractObject,
		ClusterObject,
		RuleObject,
		ACLRuleObject,
		DependencyObject,
	}
)
