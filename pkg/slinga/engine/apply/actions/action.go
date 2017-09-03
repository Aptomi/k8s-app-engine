package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type Action interface {
	object.Base
	Apply(*ActionContext) error
}
