package action

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type Base interface {
	object.Base
	Apply(*Context) error
}
