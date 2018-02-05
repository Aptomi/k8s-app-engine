package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	"sync"
)

// ServiceObject is an informational data structure with Kind and Constructor for Service
var ServiceObject = &runtime.Info{
	Kind:        "service",
	Storable:    true,
	Versioned:   true,
	Deletable:   true,
	Constructor: func() runtime.Object { return &Service{} },
}

// Service defines individual service in Aptomi. The idea is that services get defined by different teams. Those
// teams define service-specific consumption rules of how others can consume their services.
//
// Service typically consists of one or more components. Each component can either be pointer to the code (e.g.
// docker container image with metadata that needs to be started/managed) or it can be dependency on another\
// contract (which will get fulfilled by Aptomi)
type Service struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         `validate:"required"`

	// Labels is a set of labels attached to the service
	Labels map[string]string `yaml:"labels,omitempty" validate:"omitempty,labels"`

	// Components is the list of components service consists of
	Components []*ServiceComponent `validate:"dive"`

	// Lazily evaluated fields (all components topologically sorted). Use via getter
	componentsOrderedOnce sync.Once
	componentsOrderedErr  error
	componentsOrdered     []*ServiceComponent

	componentsMapOnce sync.Once
	componentsMap     map[string]*ServiceComponent
}

// ServiceComponent defines component within a service
type ServiceComponent struct {
	// Name is a user-defined component name
	Name string `validate:"identifier"`

	// Contract, if not empty, denoted that the component points to another contract as a dependency. Meaning that
	// a service needs to have another service running as its dependency (e.g. 'wordpress' service needs a 'database'
	// contract). This dependency will be fulfilled at policy resolution time.
	Contract string `yaml:"contract,omitempty" validate:"omitempty"`

	// Code, if not empty, means that component is a code that can be instantiated with certain parameters (e.g. docker
	// container image)
	Code *Code `yaml:"code,omitempty" validate:"omitempty"`

	// Discovery is a map of discovery parameters that this component exposes to other services
	Discovery util.NestedParameterMap `yaml:"discovery,omitempty" validate:"omitempty,templateNestedMap"`

	// Dependencies is cross-component dependencies within a service. Component may need other components within that
	// service to run, before it gets instantiated
	Dependencies []string `yaml:"dependencies,omitempty" validate:"dive,identifier"`
}

// Code with type and parameters, used to instantiate/update/delete component instances
type Code struct {
	// Type represents code type (e.g. aptomi/code/kubernetes-helm). It determines the plugin that will get executed for
	// for this code component
	Type string `validate:"required,codetype"`

	// Params define parameters that will be passed down to the deployment plugin. Params follow text template syntax
	// and can refer to arbitrary labels, as well as discovery parameters exposed by other components (within the
	// current service) and discovery parameters exposed by services the current service depends on
	Params util.NestedParameterMap `validate:"omitempty,templateNestedMap"`
}

// GetComponentsMap lazily initializes and returns a map of name -> component, while being thread-safe
func (service *Service) GetComponentsMap() map[string]*ServiceComponent {
	service.componentsMapOnce.Do(func() {
		// Put all components into map
		service.componentsMap = make(map[string]*ServiceComponent)
		for _, c := range service.Components {
			service.componentsMap[c.Name] = c
		}
	})
	return service.componentsMap
}

// Topologically sort components of a given service and return error if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) error {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.GetComponentsMap()[vName]
		if !exists {
			return fmt.Errorf("service '%s' has a dependency on non-existing component '%s'", service.Name, vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found or another error occurred
			if err := service.dfsComponentSort(v, colors); err != nil {
				return err
			}
		} else if vColor == 1 {
			return fmt.Errorf("component cycle detected while processing service '%s' component '%s'", service.Name, vName)
		}
	}

	service.componentsOrdered = append(service.componentsOrdered, u)
	colors[u.Name] = 2
	return nil
}

// GetComponentsSortedTopologically returns all components sorted in a topological order
// This should be thread safe
func (service *Service) GetComponentsSortedTopologically() ([]*ServiceComponent, error) {
	service.componentsOrderedOnce.Do(func() {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		for _, c := range service.Components {
			if _, ok := colors[c.Name]; !ok {
				if err := service.dfsComponentSort(c, colors); err != nil {
					service.componentsOrdered = nil
					service.componentsOrderedErr = err
					break
				}
			}
		}
	})

	return service.componentsOrdered, service.componentsOrderedErr
}
