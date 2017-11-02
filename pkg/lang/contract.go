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

// Contract is an object, which allows you to define a contract for a service, as well as a set of specific
// implementations. For example, contract can be a 'database', with specific service contexts represented
// by 'MySQL', 'MariaDB', 'SQLite'.
//
// When dependencies get declared, they always get declared on a contract (not on a specific service).
type Contract struct {
	Metadata `validate:"required"`

	// ChangeLabels defines how current set of labels will get changed/transformed in case
	// the contract gets matched
	ChangeLabels LabelOperations `yaml:"change-labels" validate:"labelOperations"`

	// Contexts contains an ordered list of contexts within a contract. When allocating an instance, Aptomi will pick
	// and instantiate the first context which matches the criteria
	Contexts []*Context `validate:"dive"`
}
