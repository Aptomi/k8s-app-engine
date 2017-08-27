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

func checkTopologicalSort(t *testing.T, serviceName string, expectedComponents []string, expectedError bool) {
	policy := LoadUnitTestsPolicy()
	service := policy.Services[serviceName]
	componentsSorted, err := service.GetComponentsSortedTopologically()
	componentsSortedStr := toStringArray(componentsSorted)
	assert.Equal(t, expectedError, err != nil, "Topological sort method (success vs. error), service: "+service.Name)
	if err == nil {
		assert.Equal(t, expectedComponents, componentsSortedStr, "Topological sort should produce correct ordering of components, service: "+service.Name)
	}
}

func TestServiceComponentsTopologicalSort(t *testing.T) {
	checkTopologicalSort(t, "kafka", []string{"component4", "component1", "component2", "component3"}, false)
	checkTopologicalSort(t, "cyclic", nil, true)
	checkTopologicalSort(t, "badcomponentdependency", nil, true)
}
