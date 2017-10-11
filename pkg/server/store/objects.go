package store

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/object"
)

var (
	Objects = []*object.Info{PolicyDataObject, RevisionDataObject, resolve.ComponentInstanceObject}
)
