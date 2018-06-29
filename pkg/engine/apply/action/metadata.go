package action

import (
	"strings"

	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Metadata is an object metadata for all state update actions
type Metadata struct {
	Kind string
	Name string
}

// NewMetadata creates new Metadata
func NewMetadata(kind string, keys ...string) *Metadata {
	keysStr := strings.Join(keys, runtime.KeySeparator)
	name := strings.Join([]string{kind, keysStr}, runtime.KeySeparator)
	return &Metadata{
		Kind: kind,
		Name: name,
	}
}

// GetKind return an action kind
func (meta *Metadata) GetKind() string {
	return meta.Kind
}

// GetName returns an action name
func (meta *Metadata) GetName() string {
	return meta.Name
}

func (meta *Metadata) String() string {
	return meta.GetName()
}
