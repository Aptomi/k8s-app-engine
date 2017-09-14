package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func toStringArray(components []*ServiceComponent) []string {
	result := []string{}
	for _, component := range components {
		result = append(result, component.Name)
	}
	return result
}

func checkTopologicalSort(t *testing.T, service *Service, expectedComponents []string, expectedError bool) {
	componentsSorted, err := service.GetComponentsSortedTopologically()
	componentsSortedStr := toStringArray(componentsSorted)
	assert.Equal(t, expectedError, err != nil, "Topological sort method (success vs. error), service: "+service.Name)
	if err == nil {
		assert.Equal(t, expectedComponents, componentsSortedStr, "Topological sort should produce correct ordering of components, service: "+service.Name)
	}
}

func TestServiceComponentsTopologicalSort(t *testing.T) {
	checkTopologicalSort(t, makeNormalService(), []string{"component4", "component1", "component2", "component3"}, false)
	checkTopologicalSort(t, makeCyclicService(), nil, true)
	checkTopologicalSort(t, makeBadComponentDependencyService(), nil, true)
}

func makeNormalService() *Service {
	return &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "normal",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:         "component1",
				Dependencies: []string{"component4"},
			},
			{
				Name:         "component2",
				Dependencies: []string{"component1"},
			},
			{
				Name:         "component3",
				Dependencies: []string{"component1", "component2"},
			},
			{
				Name:         "component4",
				Dependencies: []string{},
			},
		},
	}
}

func makeCyclicService() *Service {
	return &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "badcomponentdependency",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:         "component1",
				Dependencies: []string{"component2"},
			},
			{
				Name:         "component2",
				Dependencies: []string{"component3"},
			},
			{
				Name:         "component3",
				Dependencies: []string{"component4"},
			},
			{
				Name:         "component4",
				Dependencies: []string{"component2"},
			},
		},
	}
}

func makeBadComponentDependencyService() *Service {
	return &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "cyclic",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:         "component1",
				Dependencies: []string{"component2"},
			},
			{
				Name:         "component2",
				Dependencies: []string{"component3"},
			},
			{
				Name:         "component3",
				Dependencies: []string{"component4"},
			},
			{
				Name:         "component4",
				Dependencies: []string{"component5"},
			},
		},
	}
}
