package language

import (
	"fmt"
	"encoding/json"
	"github.com/Frostman/aptomi/pkg/slinga/language/yaml"
)

// Service defines individual service
type Service struct {
	Enabled    bool
	Name       string
	Owner      string
	Labels     *LabelOperations
	Components []*ServiceComponent

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
	Discovery    ParameterTree
	Dependencies []string
	Labels       *LabelOperations
}

// ParameterTree is a special type alias defined for freeform blocks with parameters
type ParameterTree interface{}

// Code with type and parameters, used to instantiate/update/delete component instances
type Code struct {
	Type   string
	Params ParameterTree
}

// LabelOperations defines the set of label manipulations (e.g. set/remove)
type LabelOperations map[string]map[string]string

// UnmarshalYAML is a custom unmarshaller for Service, which sets Enabled to True before unmarshalling
func (service *Service) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias Service
	instance := Alias{Enabled: true}
	if err := unmarshal(&instance); err != nil {
		return err
	}
	*service = Service(instance)
	return nil
}

// MarshalJSON marshals service component into a structure without freeform parameters, so UI doesn't fail
// See http://choly.ca/post/go-json-marshalling/
func (component *ServiceComponent) MarshalJSON() ([]byte, error) {
	type Alias ServiceComponent
	return json.Marshal(&struct {
		Code      *Code
		Discovery ParameterTree
		*Alias
	}{
		Code:      nil,
		Discovery: nil,
		Alias:     (*Alias)(component),
	})
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

// Topologically sort components of a given service and return true if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) error {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.GetComponentsMap()[vName]
		if !exists {
			return fmt.Errorf("Service %s has a dependency to non-existing component %s", service.Name, vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found or another error occured
			if err := service.dfsComponentSort(v, colors); err != nil {
				return err
			}
		} else if vColor == 1 {
			return fmt.Errorf("Component cycle detected while processing service %s", service.Name)
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

// Loads service from file
func loadServiceFromFile(fileName string) *Service {
	return yaml.LoadObjectFromFile(fileName, new(Service)).(*Service)
}
