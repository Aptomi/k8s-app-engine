package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ContractObject = &object.Info{
	Kind:        "contract",
	Constructor: func() object.Base { return &Contract{} },
}

// Contract defines a collection of individual service
type Contract struct {
	Metadata

	Owner        string
	ChangeLabels LabelOperations `yaml:"change-labels"`
	Components   []*ServiceComponent
}

// list of contexts[] -> each gets implemented via specific service
