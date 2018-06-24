package lang

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// ClaimObject is an informational data structure with Kind and Constructor for Claim
var ClaimObject = &runtime.Info{
	Kind:        "claim",
	Storable:    true,
	Versioned:   true,
	Deletable:   true,
	Constructor: func() runtime.Object { return &Claim{} },
}

// Claim is a declaration of use, defined in a form <User> needs an instance of <Contract> with
// specified set of <Labels>. It allows users to request contracts, which will translate into instantiation of
// complete service instances in the cloud
type Claim struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         `validate:"required"`

	// User is a user name for a user, who requested this claim.
	User string `validate:"required"`

	// Contract that is being requested. It can be in form of 'contractName', referring to contract within
	// current namespace. Or it can be in form of 'namespace/contractName', referring to contract in a different
	// namespace.
	Contract string `validate:"required"`

	// Labels which are provided by the user.
	Labels map[string]string `yaml:"labels,omitempty" validate:"omitempty,labels"`
}
