package store

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

func (s *defaultStore) GetRevision(object.Generation) (*lang.Policy, error) {
	return nil, nil
}

func (s *defaultStore) UpdateRevision() error {
	return nil
}
