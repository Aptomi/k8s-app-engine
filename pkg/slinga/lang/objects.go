package lang

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var (
	Objects = []*object.Info{
		ServiceObject,
		ContractObject,
		ClusterObject,
		RuleObject,
		DependencyObject,
	}
)
