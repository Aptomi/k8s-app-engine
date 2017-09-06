package action

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type Action interface {
	object.Base
	Apply(*Context) error
}
