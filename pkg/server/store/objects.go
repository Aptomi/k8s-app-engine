package store

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/object"
)

var (
	// Objects is the list of object.Info for all server store objects used
	Objects = []*object.Info{PolicyDataObject, RevisionDataObject, resolve.ComponentInstanceObject}
)
