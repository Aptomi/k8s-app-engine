package action

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

// Metadata is an object metadata for all state update actions
type Metadata struct {
	Key      string
	Kind     string
	Revision object.Generation
}

// NewMetadata creates new Metadata
func NewMetadata(revision object.Generation, kind string, keys ...string) *Metadata {
	keysStr := strings.Join(keys, object.KeySeparator)
	return &Metadata{
		Key:      strings.Join([]string{revision.String(), kind, keysStr}, object.KeySeparator),
		Kind:     kind,
		Revision: revision,
	}
}

// GetKey returns an object key
func (meta *Metadata) GetKey() string {
	return meta.Key
}

// GetNamespace returns a namespace for an action (it's always a system namespace)
func (meta *Metadata) GetNamespace() string {
	return object.SystemNS
}

// GetKind returns an object kind
func (meta *Metadata) GetKind() string {
	return meta.Kind
}

// GetGeneration returns a generation for action (it's always zero as actions are not versioned)
func (meta *Metadata) GetGeneration() object.Generation {
	// we aren't storing action versions
	return 0
}

// SetGeneration for an action (not needed)
func (meta *Metadata) SetGeneration(generation object.Generation) {
	panic("Action is not a versioned object")
}
