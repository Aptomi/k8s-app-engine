package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// ContractObject is an informational data structure with Kind and Constructor for Contract
var ContractObject = &object.Info{
	Kind:        "contract",
	Versioned:   true,
	Constructor: func() object.Base { return &Contract{} },
}

// Contract defines a contract for service usage
type Contract struct {
	Metadata

	// ChangeLabels contains change label actions in the policy
	ChangeLabels LabelOperations `yaml:"change-labels"`

	// Contexts contains an ordered list of contexts within a contract
	Contexts []*Context
}

// FindContextByName finds a context by name
func (contract *Contract) FindContextByName(contextName string) *Context {
	for _, context := range contract.Contexts {
		if context.Name == contextName {
			return context
		}
	}
	return nil
}
