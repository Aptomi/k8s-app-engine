package action

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	"strings"
)

// Metadata is an object metadata for all state update actions
type Metadata struct {
	Name string
}

// NewMetadata creates new Metadata
func NewMetadata(kind string, keys ...string) *Metadata {
	keysStr := strings.Join(keys, runtime.KeySeparator)
	name := strings.Join([]string{kind, keysStr}, runtime.KeySeparator)
	return &Metadata{
		Name: name,
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
