package lang

import (
	"fmt"
	"sync"

	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// TypeBundle is an informational data structure with Kind and Constructor for Bundle
var TypeBundle = &runtime.TypeInfo{
	Kind:        "bundle",
	Storable:    true,
	Versioned:   true,
	Constructor: func() runtime.Object { return &Bundle{} },
}

// Bundle defines individual bundle in Aptomi. The idea is that bundles get defined by different teams. Those
// teams define bundle-specific consumption rules of how others can consume their bundles.
//
// Bundle typically consists of one or more components. Each component can either be pointer to the code (e.g.
// docker container image with metadata that needs to be started/managed) or it can be dependency on another
// service (which will get fulfilled by Aptomi)
type Bundle struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         `validate:"required"`

	// Labels is a set of labels attached to the bundle
	Labels map[string]string `yaml:"labels,omitempty" validate:"omitempty,labels"`

	// Components is the list of components bundle consists of
	Components []*BundleComponent `validate:"dive"`

	// Lazily evaluated fields (all components topologically sorted). Use via getter
	componentsOrderedOnce sync.Once
	componentsOrderedErr  error
	componentsOrdered     []*BundleComponent

	componentsMapOnce sync.Once
	componentsMap     map[string]*BundleComponent
}

// BundleComponent defines component within a bundle
type BundleComponent struct {
	// Name is a user-defined component name
	Name string `validate:"identifier"`

	// Criteria - if it gets evaluated to true during policy resolution, then component will be included
	// into the bundle. It's an optional field, so if it's nil then it is considered to be true automatically
	Criteria *Criteria `yaml:",omitempty" validate:"omitempty"`

	// Service, if not empty, denoted that the component points to another service. Meaning that
	// a bundle needs to have another service instantiated and running (e.g. 'wordpress' bundle needs a 'database'
	// service). This will be fulfilled at policy resolution time.
	Service string `yaml:"service,omitempty" validate:"omitempty"`

	// Code, if not empty, means that component is a code that can be instantiated with certain parameters (e.g. docker
	// container image)
	Code *Code `yaml:"code,omitempty" validate:"omitempty"`

	// Discovery is a map of discovery parameters that this component exposes to other bundles
	Discovery util.NestedParameterMap `yaml:"discovery,omitempty" validate:"omitempty,templateNestedMap"`

	// Dependencies represent cross-component dependencies within a given bundle. Component may need other components
	// within that bundle to exist, before it gets instantiated
	Dependencies []string `yaml:"dependencies,omitempty" validate:"dive,identifier"`
}

// Code with type and parameters, used to instantiate/update/delete component instances
type Code struct {
	// Type represents code type (e.g. "helm"). It determines the plugin that will get executed for
	// for this code component
	Type string `validate:"required,codetype"`

	// Params define parameters that will be passed down to the deployment plugin. Params follow text template syntax
	// and can refer to arbitrary labels, as well as discovery parameters exposed by other components (within the
	// current bundle) and discovery parameters exposed by bundles the current bundle depends on
	Params util.NestedParameterMap `validate:"omitempty,templateNestedMap"`
}

// Matches checks if component criteria is satisfied
func (component *BundleComponent) Matches(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if component.Criteria == nil {
		return true, nil
	}
	return component.Criteria.allows(params, cache)
}

// GetComponentsMap lazily initializes and returns a map of name -> component, while being thread-safe
func (bundle *Bundle) GetComponentsMap() map[string]*BundleComponent {
	bundle.componentsMapOnce.Do(func() {
		// Put all components into map
		bundle.componentsMap = make(map[string]*BundleComponent)
		for _, c := range bundle.Components {
			bundle.componentsMap[c.Name] = c
		}
	})
	return bundle.componentsMap
}

// Topologically sort components of a given bundle and return error if there is a cycle detected
func (bundle *Bundle) dfsComponentSort(u *BundleComponent, colors map[string]int) error {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := bundle.GetComponentsMap()[vName]
		if !exists {
			return fmt.Errorf("bundle '%s' has a claim on non-existing component '%s'", bundle.Name, vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found or another error occurred
			if err := bundle.dfsComponentSort(v, colors); err != nil {
				return err
			}
		} else if vColor == 1 {
			return fmt.Errorf("component cycle detected while processing bundle '%s' component '%s'", bundle.Name, vName)
		}
	}

	bundle.componentsOrdered = append(bundle.componentsOrdered, u)
	colors[u.Name] = 2
	return nil
}

// GetComponentsSortedTopologically returns all components sorted in a topological order
// This should be thread safe
func (bundle *Bundle) GetComponentsSortedTopologically() ([]*BundleComponent, error) {
	bundle.componentsOrderedOnce.Do(func() {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		for _, c := range bundle.Components {
			if _, ok := colors[c.Name]; !ok {
				if err := bundle.dfsComponentSort(c, colors); err != nil {
					bundle.componentsOrdered = nil
					bundle.componentsOrderedErr = err
					break
				}
			}
		}
	})

	return bundle.componentsOrdered, bundle.componentsOrderedErr
}
