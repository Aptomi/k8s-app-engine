package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ContractObject = &object.Info{
	Kind:        "contract",
	Constructor: func() object.Base { return &Contract{} },
}

// Contract defines a contract for service usage
type Contract struct {
	Metadata

	// List of contexts
	Contexts []*Context
}
