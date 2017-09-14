package action

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"strings"
)

type Metadata struct {
	Key      string
	Kind     string
	Revision object.Generation
}

func NewMetadata(revision object.Generation, kind string, keys ...string) *Metadata {
	keysStr := strings.Join(keys, object.KeySeparator)
	return &Metadata{
		Key:      strings.Join([]string{revision.String(), kind, keysStr}, object.KeySeparator),
		Kind:     kind,
		Revision: revision,
	}
}

func (meta *Metadata) GetKey() string {
	return meta.Key
}

func (meta *Metadata) GetNamespace() string {
	return object.SystemNS
}

func (meta *Metadata) GetKind() string {
	return meta.Kind
}

func (meta *Metadata) GetGeneration() object.Generation {
	// we aren't storing action versions
	return 0
}

func (meta *Metadata) SetGeneration(generation object.Generation) {
	panic("ComponentInstance isn't a versioned object")
}
