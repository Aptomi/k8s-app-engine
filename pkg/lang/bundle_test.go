package lang

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/stretchr/testify/assert"
)

func TestBundleComponentCriteria(t *testing.T) {
	params := expression.NewParams(
		map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
		nil,
	)
	cache := expression.NewCache()
	checkMatch(t, true, &BundleComponent{Criteria: &Criteria{RequireAll: []string{"param1 == 'value1' && param2 == 'value2'"}}}, params, cache)
	checkMatch(t, false, &BundleComponent{Criteria: &Criteria{RequireAll: []string{"param1 == 'somevalue'"}}}, params, cache)
	checkMatch(t, true, &BundleComponent{}, params, cache)
}

func checkMatch(t *testing.T, expectedValue bool, component *BundleComponent, params *expression.Parameters, cache *expression.Cache) {
	t.Helper()
	res, err := component.Matches(params, cache)
	if err != nil {
		t.Fail()
		return
	}
	assert.Equal(t, expectedValue, res, "Bundle component criteria match should produce correct result. Expected %t for criteria: %+v", expectedValue, component.Criteria)
}

func TestBundleComponentsTopologicalSort(t *testing.T) {
	checkTopologicalSort(t, makeNormalBundle(), []string{"component4", "component1", "component2", "component3"}, false)
	checkTopologicalSort(t, makeCyclicBundle(), nil, true)
	checkTopologicalSort(t, makeBadComponentClaimBundle(), nil, true)
}
func toStringArray(components []*BundleComponent) []string {
	result := []string{}
	for _, component := range components {
		result = append(result, component.Name)
	}
	return result
}

func checkTopologicalSort(t *testing.T, bundle *Bundle, expectedComponents []string, expectedError bool) {
	t.Helper()
	componentsSorted, err := bundle.GetComponentsSortedTopologically()
	componentsSortedStr := toStringArray(componentsSorted)
	assert.Equal(t, expectedError, err != nil, "Topological sort method (success vs. error), bundle: %s", bundle.Name)
	if err == nil {
		assert.Equal(t, expectedComponents, componentsSortedStr, "Topological sort should produce correct ordering of components, bundle: %s", bundle.Name)
	}
}

func makeNormalBundle() *Bundle {
	return &Bundle{
		TypeKind: TypeBundle.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "normal",
		},
		Components: []*BundleComponent{
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

func makeCyclicBundle() *Bundle {
	return &Bundle{
		TypeKind: TypeBundle.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "badcomponentclaim",
		},
		Components: []*BundleComponent{
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

func makeBadComponentClaimBundle() *Bundle {
	return &Bundle{
		TypeKind: TypeBundle.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "cyclic",
		},
		Components: []*BundleComponent{
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
