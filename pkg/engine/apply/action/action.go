package action

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// Base interface for all actions which perform actual state updates
type Base interface {
	runtime.Storable
	Apply(*Context) error
	DescribeChanges() util.NestedParameterMap
}
