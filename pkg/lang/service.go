package lang

import (
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// ServiceObject is an informational data structure with Kind and Constructor for Service
var ServiceObject = &runtime.Info{
	Kind:        "service",
	Storable:    true,
	Versioned:   true,
	Deletable:   true,
	Constructor: func() runtime.Object { return &Service{} },
}

// Service is an object, which allows you to define a service for a bundle, as well as a set of specific
// implementations. For example, service can be a 'database', with specific bundle contexts implemented
// by 'MySQL', 'MariaDB', 'SQLite'.
//
// When claims get declared, they always get declared on a service (not on a specific bundle).
type Service struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         `validate:"required"`

	// ChangeLabels defines how current set of labels will get changed/transformed in case
	// the service gets matched
	ChangeLabels LabelOperations `yaml:"change-labels,omitempty" validate:"labelOperations"`

	// Contexts contains an ordered list of contexts within a service. When allocating an instance, Aptomi will pick
	// and instantiate the first context which matches the criteria
	Contexts []*Context `validate:"dive"`
}

// Context represents a single context within a service.
// It's essentially a bundle instance for a given of class of use cases, a given set of consumers, etc.
type Context struct {
	// Name defines context name in the policy
	Name string `validate:"identifier"`

	// Criteria - if it gets evaluated to true during policy resolution, then service
	// will get fulfilled by allocating this context. It's an optional field, so if it's nil then
	// it is considered to be evaluated to true automatically
	Criteria *Criteria `validate:"omitempty"`

	// ChangeLabels defines how current set of labels will get changed/transformed in case
	// the context gets matched
	ChangeLabels LabelOperations `yaml:"change-labels,omitempty" validate:"labelOperations"`

	// Allocation defines how the context will get allocated (which bundle to allocate and which unique key to use)
	Allocation *Allocation `validate:"required"`
}

// Allocation determines which bundle should be allocated for by the given context
// and which additional keys should be added to component instance key
type Allocation struct {
	// Bundle defined which bundle to allocated. It can be in form of 'bundleName', referring to bundle within
	// current namespace. Or it can be in form of 'namespace/bundleName', referring to bundle in a different
	// namespace
	Bundle string `validate:"required"`

	// Keys define a set of unique keys that define this allocation. If keys are not defined, then allocation will
	// always correspond to a single instance. If keys are defined, it will allow to create different bundle instances
	// based on labels. Different keys values resolved during policy processing will result in different bundle
	// instances created by Aptomi. For example, if key is set to {{.User.Labels.team}}, it will get dynamically
	// resolved into a user's team name. And, since users from different teams will have different keys, every team
	// will get their own bundle instance from Aptomi
	Keys []string `yaml:"keys,omitempty" validate:"dive,template"`
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if context.Criteria == nil {
		return true, nil
	}
	return context.Criteria.allows(params, cache)
}

// ResolveKeys resolves dynamic allocation keys, which later get added to component instance key
func (context *Context) ResolveKeys(params *template.Parameters, cache *template.Cache) ([]string, error) {
	if cache == nil {
		cache = template.NewCache()
	}
	// Resolve allocation keys (they can be dynamic, depending on user labels)
	result := []string{}
	for _, key := range context.Allocation.Keys {
		keyResolved, err := cache.Evaluate(key, params)
		if err != nil {
			return nil, err
		}
		result = append(result, keyResolved)
	}
	return result, nil
}
