package action

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// Base interface for all actions which perform actual state updates
type Base interface {
	runtime.Storable
	AfterCreated(*resolve.PolicyResolution)
	Apply(*Context) error
	DescribeChanges() util.NestedParameterMap
}
