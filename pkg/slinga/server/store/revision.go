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

// RevisionData is a "milestone" in applying
type RevisionData struct {
	lang.Metadata

	// Policy represents generation of the corresponding policy
	Policy object.Generation
}

func (s *defaultStore) GetRevision(object.Generation) (*RevisionData, error) {
	return nil, nil
}

func (s *defaultStore) NextRevision(policyGen object.Generation) (*RevisionData, error) {
	return &RevisionData{
		Metadata: lang.Metadata{
			Namespace: object.SystemNS,
			Kind:      RevisionDataObject.Kind,
			Name:      RevisionName,

			// todo find next generation
			Generation: 0,
		},
		Policy: policyGen,
	}, nil
}
