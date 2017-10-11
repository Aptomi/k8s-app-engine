package action

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// Base interface for all actions which perform actual state updates
type Base interface {
	object.Base
	Apply(*Context) error
}
