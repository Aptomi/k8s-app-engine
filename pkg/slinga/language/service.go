package language

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

var ServiceObject = &ObjectInfo{
	Kind("service"),
	func() BaseObject { return &Service{} },
}

// Service defines individual service
type Service struct {
	Metadata

	Owner        string
	ChangeLabels LabelOperations `yaml:"change-labels"`
	Components   []*ServiceComponent

	// Lazily evaluated field (all components topologically sorted). Use via getter
	componentsOrdered []*ServiceComponent

	// Lazily evaluated field. Use via getter
	componentsMap map[string]*ServiceComponent
}

// ServiceComponent defines component within a service
type ServiceComponent struct {
	Name         string
	Service      string
	Code         *Code
	Discovery    NestedParameterMap
	Dependencies []string
	ChangeLabels LabelOperations `yaml:"change-labels"`
}

// Code with type and parameters, used to instantiate/update/delete component instances
type Code struct {
	Type   string
	Params NestedParameterMap
}

// GetComponentsMap lazily initializes and returns a map of name -> component
func (service *Service) GetComponentsMap() map[string]*ServiceComponent {
	if service.componentsMap == nil {
		// Put all components into map
		service.componentsMap = make(map[string]*ServiceComponent)
		for _, c := range service.Components {
			service.componentsMap[c.Name] = c
		}
	}
	return service.componentsMap
}

// Topologically sort components of a given service and return error if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) error {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.GetComponentsMap()[vName]
		if !exists {
			return fmt.Errorf("Service '%s' has a dependency on non-existing component '%s'", service.Name, vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found or another error occured
			if err := service.dfsComponentSort(v, colors); err != nil {
				return err
			}
		} else if vColor == 1 {
			return fmt.Errorf("Component cycle detected while processing service '%s' component '%s'", service.Name, vName)
		}
	}

	service.componentsOrdered = append(service.componentsOrdered, u)
	colors[u.Name] = 2
	return nil
}

// GetComponentsSortedTopologically returns all components sorted in a topological order
func (service *Service) GetComponentsSortedTopologically() ([]*ServiceComponent, error) {
	if service.componentsOrdered == nil {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		for _, c := range service.Components {
			if _, ok := colors[c.Name]; !ok {
				if err := service.dfsComponentSort(c, colors); err != nil {
					return nil, err
				}
			}
		}
	}

	return service.componentsOrdered, nil
}
