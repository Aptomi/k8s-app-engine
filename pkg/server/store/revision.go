package store

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
)

// RevisionName is the name of the only revision that exists in DB (but with many generations)
const RevisionName = "revision"

// RevisionDataObject is object.Info for RevisionData
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

	Progress revisionProgress
}

// GetRevision returns RevisionData for specified generation
func (s *DefaultStore) GetRevision(gen object.Generation) (*RevisionData, error) {
	dataObj, err := s.store.GetByName(object.SystemNS, RevisionDataObject.Kind, RevisionName, gen)
	if err != nil {
		return nil, err
	}
	if dataObj == nil {
		return nil, nil
	}
	data, ok := dataObj.(*RevisionData)
	if !ok {
		return nil, fmt.Errorf("unexpected type while getting RevisionData from DB")
	}
	return data, nil
}

// NextRevision returns new RevisionData for specified policy generation
func (s *DefaultStore) NextRevision(policyGen object.Generation) (*RevisionData, error) {
	currRevision, err := s.GetRevision(object.LastGen)
	if err != nil {
		return nil, fmt.Errorf("error while geting current revision: %s", err)
	}
	var gen object.Generation
	if currRevision == nil {
		gen = object.FirstGen
	} else {
		gen = currRevision.GetGeneration().Next()
	}

	return &RevisionData{
		Metadata: lang.Metadata{
			Namespace:  object.SystemNS,
			Kind:       RevisionDataObject.Kind,
			Name:       RevisionName,
			Generation: gen,
		},
		Policy: policyGen,
	}, nil
}

// SaveRevision saves specified RevisionData into the store
func (s *DefaultStore) SaveRevision(revision *RevisionData) error {
	_, err := s.store.Save(revision)
	if err != nil {
		return fmt.Errorf("error while saving revision: %s", err)
	}

	return nil
}
