package lang

import (
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceComponentCriteria(t *testing.T) {
	params := expression.NewParams(
		map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
		nil,
	)
	cache := expression.NewCache()
	checkMatch(t, true, &ServiceComponent{Criteria: &Criteria{RequireAll: []string{"param1 == 'value1' && param2 == 'value2'"}}}, params, cache)
	checkMatch(t, false, &ServiceComponent{Criteria: &Criteria{RequireAll: []string{"param1 == 'somevalue'"}}}, params, cache)
	checkMatch(t, true, &ServiceComponent{}, params, cache)
}

func checkMatch(t *testing.T, expectedValue bool, component *ServiceComponent, params *expression.Parameters, cache *expression.Cache) {
	t.Helper()
	res, err := component.Matches(params, cache)
	if err != nil {
		t.Fail()
		return
	}
	assert.Equal(t, expectedValue, res, "Service component criteria match should produce correct result. Expected %t for criteria: %+v", expectedValue, component.Criteria)
}

func TestServiceComponentsTopologicalSort(t *testing.T) {
	checkTopologicalSort(t, makeNormalService(), []string{"component4", "component1", "component2", "component3"}, false)
	checkTopologicalSort(t, makeCyclicService(), nil, true)
	checkTopologicalSort(t, makeBadComponentDependencyService(), nil, true)
}
func toStringArray(components []*ServiceComponent) []string {
	result := []string{}
	for _, component := range components {
		result = append(result, component.Name)
	}
	return result
}

func checkTopologicalSort(t *testing.T, service *Service, expectedComponents []string, expectedError bool) {
	t.Helper()
	componentsSorted, err := service.GetComponentsSortedTopologically()
	componentsSortedStr := toStringArray(componentsSorted)
	assert.Equal(t, expectedError, err != nil, "Topological sort method (success vs. error), service: "+service.Name)
	if err == nil {
		assert.Equal(t, expectedComponents, componentsSortedStr, "Topological sort should produce correct ordering of components, service: "+service.Name)
	}
}

func makeNormalService() *Service {
	return &Service{
		TypeKind: ServiceObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "normal",
		},
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
		TypeKind: ServiceObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "badcomponentdependency",
		},
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
		TypeKind: ServiceObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "cyclic",
		},
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
