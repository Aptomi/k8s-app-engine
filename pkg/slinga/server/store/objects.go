package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var (
	Objects = []*object.Info{PolicyDataObject, RevisionDataObject, resolve.ComponentInstanceObject}
)
