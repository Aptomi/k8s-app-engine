package action

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	"strings"
)

// Metadata is an object metadata for all state update actions
type Metadata struct {
	Name     string
	Revision runtime.Generation
}

// NewMetadata creates new Metadata
func NewMetadata(revision runtime.Generation, kind string, keys ...string) *Metadata {
	keysStr := strings.Join(keys, runtime.KeySeparator)
	name := strings.Join([]string{revision.String(), kind, keysStr}, runtime.KeySeparator)
	return &Metadata{
		Name:     name,
		Revision: revision,
	}
}

// GetName returns an action name
func (meta *Metadata) GetName() string {
	return meta.Name
}

// GetNamespace returns a namespace for an action (it's always a system namespace)
func (meta *Metadata) GetNamespace() string {
	return runtime.SystemNS
}

func (meta *Metadata) String() string {
	return meta.GetName()
}
