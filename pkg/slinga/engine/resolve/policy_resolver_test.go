package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPolicyResolverAndResolvedData(t *testing.T) {
	policy, usageState := loadPolicyAndResolve(t)
	resolvedData := usageState.ResolvedData

	// Check that policy resolution finished correctly
	assert.Equal(t, 16, len(resolvedData.ComponentProcessingOrder), "Policy usage should have correct number of entries")

	// Resolution for test context
	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policy, usageState)
	assert.Equal(t, 1, len(kafkaTest.DependencyIds), "One dependency should be resolved with access to test")
	assert.True(t, policy.Dependencies.DependenciesByID["dep_id_1"].Resolved, "Only Alice should have access to test")

	// Resolution for prod context
	kafkaProd := getInstanceByParams(t, "kafka", "prod-low", []string{"team-platform_services", "true"}, "component2", policy, usageState)
	assert.Equal(t, 1, len(kafkaProd.DependencyIds), "One dependency should be resolved with access to prod")
	assert.Equal(t, "2", policy.Dependencies.DependenciesByID["dep_id_2"].UserID, "Only Bob should have access to prod (Carol is compromised)")
}

func TestPolicyResolverAndUnresolvedData(t *testing.T) {
	policy, _ := loadPolicyAndResolve(t)

	// Dave dependency on kafka should not be resolved
	daveOnKafkaDependency := policy.Dependencies.DependenciesByID["dep_id_4"]
	assert.False(t, daveOnKafkaDependency.Resolved, "Partial matching is broken. User has access to kafka, but not to zookeeper that kafka depends on. This should not be resolved successfully")
}

func TestPolicyResolverLabelProcessing(t *testing.T) {
	policy, usageState := loadPolicyAndResolve(t)

	// Check labels for Bob's dependency
	key := policy.Dependencies.DependenciesByID["dep_id_2"].ServiceKey
	serviceInstance := getInstanceInternal(t, key, usageState.ResolvedData)
	labels := serviceInstance.CalculatedLabels.Labels
	assert.Equal(t, "yes", labels["important"], "Label 'important=yes' should be carried from dependency all the way through the policy")
	assert.Equal(t, "true", labels["prod-low-ctx"], "Label 'prod-low-ctx=true' should be added on context match")
	assert.Equal(t, "", labels["some-label-to-be-removed"], "Label 'some-label-to-be-removed' should be removed on context match")
	assert.Equal(t, "true", labels["prod-low-alloc"], "Label 'prod-low-alloc=true' should be added on allocation match")
}

func TestPolicyResolverCodeAndDiscoveryParamsEval(t *testing.T) {
	policy, usageState := loadPolicyAndResolve(t)

	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policy, usageState)

	// Check that code parameters evaluate correctly
	assert.Equal(t, "zookeeper-test-platform-services-component2", kafkaTest.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")

	// Check that discovery parameters evaluate correctly
	assert.Equal(t, "kafka-kafka-test-platform-services-component2-url", kafkaTest.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// Check that nested parameters evaluate correctly
	for i := 1; i <= 5; i++ {
		assert.Equal(t, "value"+strconv.Itoa(i), kafkaTest.CalculatedCodeParams.GetNestedMap("data" + strconv.Itoa(i)).GetNestedMap("param")["name"], "Nested code parameters should be calculated correctly")
	}
}

func TestPolicyResolverDependencyWithNonExistingUser(t *testing.T) {
	policy := loadUnitTestsPolicy()

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:  "non-existing-user-123456789",
		Service: "newservice",
	}
	policy.AddObject(dependency)

	// policy with missing user should be resolved successfully
	resolvePolicy(t, policy, ResSuccess, "")

	// dependency should be just skipped
	assert.False(t, dependency.Resolved, "Dependency should not be resolved")
}

func TestPolicyResolverDependencyWithNonExistingService(t *testing.T) {
	policy := loadUnitTestsPolicy()

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:  "4",
		Service: "non-existing-service-123456789",
	}
	policy.AddObject(dependency)

	// policy with missing service should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Unable to find service definition")
}

func TestPolicyResolverInvalidContextCriteria(t *testing.T) {
	policy := loadUnitTestsPolicy()

	service := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Owner: "1",
	}
	policy.AddObject(service)

	context := &Context{
		Metadata: Metadata{
			Kind:      ContextObject.Kind,
			Namespace: "main",
			Name:      "special-invalid-context-require-any",
		},
		Criteria: &Criteria{
			RequireAll: []string{"service.Name=='xyz'"},
			RequireAny: []string{"specialname + '123')((("},
		},
	}
	policy.AddObject(context)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:  "7",
		Service: "xyz",
	}
	policy.AddObject(dependency)

	// policy with invalid context should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Unable to compile expression")
}

func TestPolicyResolverInvalidContextKeys(t *testing.T) {
	policy := loadUnitTestsPolicy()

	service := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Owner: "1",
	}
	policy.AddObject(service)

	context := &Context{
		Metadata: Metadata{
			Kind:      ContextObject.Kind,
			Namespace: "main",
			Name:      "special-invalid-context-keys",
		},
		Criteria: &Criteria{
			RequireAll: []string{"service.Name=='xyz'"},
		},
		Allocation: &struct {
			Keys []string
		}{
			Keys: []string{
				"wowowow {{{{.......",
			},
		},
	}
	policy.AddObject(context)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:  "7",
		Service: "xyz",
	}
	policy.AddObject(dependency)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Error while resolving allocation keys")
}

func TestPolicyResolverInvalidServiceWithoutOwner(t *testing.T) {
	policy := loadUnitTestsPolicy()

	service := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
	}
	policy.AddObject(service)

	context := &Context{
		Metadata: Metadata{
			Kind:      ContextObject.Kind,
			Namespace: "main",
			Name:      "special-invalid-context-keys",
		},
		Criteria: &Criteria{
			RequireAll: []string{"service.Name=='xyz'"},
		},
		Allocation: &struct {
			Keys []string
		}{
			Keys: []string{
				"wowowow {{{{.......",
			},
		},
	}
	policy.AddObject(context)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:  "7",
		Service: "xyz",
	}
	policy.AddObject(dependency)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Unable to find service owner")
}

func TestPolicyResolverInvalidRuleCriteria(t *testing.T) {
	policy := loadUnitTestsPolicy()

	rule := &Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "special-invalid-rule-require-all",
		},
		FilterServices: &ServiceFilter{
			Cluster: &Criteria{
				RequireAll: []string{"specialname + '123')((("},
			},
			Labels: &Criteria{
				RequireAll: []string{"specialname + '123')((("},
			},
			User: &Criteria{
				RequireAll: []string{"specialname + '123')((("},
			},
		},
		Actions: []*Action{
			{"dependency", "forbid"},
		},
	}
	policy.AddObject(rule)

	// policy with invalid rule should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Unable to compile expression")
}

func TestPolicyResolverConflictingCodeParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	service := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component",
				Code: &Code{
					Type: "aptomi/code/unittests",
					Params: util.NestedParameterMap{
						"address": "{{ .Labels.deplabel }}",
					},
				},
			},
		},
	}
	policy.AddObject(service)

	context := &Context{
		Metadata: Metadata{
			Kind:      ContextObject.Kind,
			Namespace: "main",
			Name:      "xyz-context",
		},
		Criteria: &Criteria{
			RequireAll: []string{"service.Name=='xyz'"},
		},
	}
	policy.AddObject(context)

	dependency1 := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new_1",
		},
		UserID:  "7",
		Service: "xyz",
		Labels: map[string]string{
			"deplabel": "1",
		},
	}
	policy.AddObject(dependency1)

	dependency2 := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new_2",
		},
		UserID:  "7",
		Service: "xyz",
		Labels: map[string]string{
			"deplabel": "2",
		},
	}
	policy.AddObject(dependency2)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Conflicting code parameters")
}

func TestPolicyResolverConflictingDiscoveryParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	service := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component",
				Discovery: util.NestedParameterMap{
					"address": "{{ .Labels.deplabel }}",
				},
			},
		},
	}
	policy.AddObject(service)

	context := &Context{
		Metadata: Metadata{
			Kind:      ContextObject.Kind,
			Namespace: "main",
			Name:      "xyz-context",
		},
		Criteria: &Criteria{
			RequireAll: []string{"service.Name=='xyz'"},
		},
	}
	policy.AddObject(context)

	dependency1 := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new_1",
		},
		UserID:  "7",
		Service: "xyz",
		Labels: map[string]string{
			"deplabel": "1",
		},
	}
	policy.AddObject(dependency1)

	dependency2 := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new_2",
		},
		UserID:  "7",
		Service: "xyz",
		Labels: map[string]string{
			"deplabel": "2",
		},
	}
	policy.AddObject(dependency2)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Conflicting discovery parameters")
}
