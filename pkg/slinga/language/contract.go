package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ContractObject = &object.Info{
	Kind:        "contract",
	Versioned: true,
	Constructor: func() object.Base { return &Contract{} },
}

// Contract defines a contract for service usage
type Contract struct {
	Metadata

	// Label changes
	ChangeLabels LabelOperations `yaml:"change-labels"`

	// List of contexts
	Contexts []*Context
}

// Returns the context by name
func (contract *Contract) FindContextByName(contextName string) *Context {
	for _, context := range contract.Contexts {
		if context.Name == contextName {
			return context
		}
	}
	return nil
}
