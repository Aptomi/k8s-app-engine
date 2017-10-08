package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

const RevisionName = "revision"

var RevisionDataObject = &object.Info{
	Kind:        "revision",
	Versioned:   true,
	Constructor: func() object.Base { return &RevisionData{} },
}

type RevisionData struct {
	lang.Metadata


}

func (s *defaultStore) GetRevision(object.Generation) (*RevisionData, error) {
	return nil, nil
}

func (s *defaultStore) NextRevision() (*RevisionData, error) {
	return &RevisionData{lang.Metadata{
		Namespace: object.SystemNS,
		Kind:      RevisionDataObject.Kind,
		Name:      RevisionName,

		// todo find next generation
		Generation: 0,
	}}, nil
}
