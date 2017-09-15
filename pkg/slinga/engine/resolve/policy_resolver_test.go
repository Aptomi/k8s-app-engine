package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPolicyResolverAndResolvedData(t *testing.T) {
	policy, resolution := loadPolicyAndResolve(t)

	// Check that policy resolution finished correctly
	assert.Equal(t, 12, len(resolution.ComponentInstanceMap), "Policy resolution data should have correct number of entries")

	// Resolution for test context
	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policy, resolution)
	assert.Equal(t, 1, len(kafkaTest.DependencyIds), "One dependency should be resolved with access to test")
	assert.NotEmpty(t, resolution.DependencyInstanceMap["dep_id_1"], "Only Alice should have access to test")

	// Resolution for prod context
	kafkaProd := getInstanceByParams(t, "kafka", "prod-low", []string{"team-platform_services", "true"}, "component2", policy, resolution)
	assert.Equal(t, 1, len(kafkaProd.DependencyIds), "One dependency should be resolved with access to prod")
	assert.Equal(t, "2", policy.Dependencies.DependenciesByID["dep_id_2"].UserID, "Only Bob should have access to prod (Carol is compromised)")
}

func TestPolicyResolverAndUnresolvedData(t *testing.T) {
	_, resolution := loadPolicyAndResolve(t)

	// Dave dependency on kafka should not be resolved
	assert.Empty(t, resolution.DependencyInstanceMap["dep_id_4"], "Partial matching is broken. User has access to kafka, but not to zookeeper that kafka depends on. This should not be resolved successfully")
}

func TestPolicyResolverLabelProcessing(t *testing.T) {
	_, resolution := loadPolicyAndResolve(t)

	// Check labels for Bob's dependency
	key := resolution.DependencyInstanceMap["dep_id_2"]
	serviceInstance := getInstanceInternal(t, key, resolution)
	labels := serviceInstance.CalculatedLabels.Labels
	assert.Equal(t, "yes", labels["important"], "Label 'important=yes' should be carried from dependency all the way through the policy")
	assert.Equal(t, "true", labels["prod-low-ctx"], "Label 'prod-low-ctx=true' should be added on context match")
	assert.Equal(t, "", labels["some-label-to-be-removed"], "Label 'some-label-to-be-removed' should be removed on context match")
	assert.Equal(t, "true", labels["prod-low-alloc"], "Label 'prod-low-alloc=true' should be added on allocation match")
}

func TestPolicyResolverCodeAndDiscoveryParamsEval(t *testing.T) {
	policy, resolution := loadPolicyAndResolve(t)

	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policy, resolution)

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
		UserID:   "non-existing-user-123456789",
		Contract: "newcontract",
	}
	policy.AddObject(dependency)

	// dependency referring to non-existing user should not trigger a critical error
	resolution := resolvePolicy(t, policy, ResSuccess, "")

	// dependency should be just skipped
	assert.Empty(t, resolution.DependencyInstanceMap["dep_id_new"], "Dependency should not be resolved")
}

func TestPolicyResolverDependencyWithNonExistingContract(t *testing.T) {
	policy := loadUnitTestsPolicy()

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "4",
		Contract: "non-existing-contract-123456789",
	}
	policy.AddObject(dependency)

	// dependency referring to non-existing contract should not trigger a critical error
	resolution := resolvePolicy(t, policy, ResSuccess, "")

	// dependency should be just skipped
	assert.Empty(t, resolution.DependencyInstanceMap["dep_id_new"], "Dependency should not be resolved")
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

	contract := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Contexts: []*Context{{
			Name: "special-invalid-context-require-any",
			Criteria: &Criteria{
				RequireAll: []string{"true"},
				RequireAny: []string{"specialname + '123')((("},
			},
			Allocation: &Allocation{
				Service: "xyz",
			},
		}},
	}
	policy.AddObject(contract)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "xyz",
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

	contract := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Contexts: []*Context{{
			Name: "special-invalid-context-keys",
			Allocation: &Allocation{
				Service: "xyz",
				Keys: []string{
					"wowowow {{{{.......",
				},
			},
		}},
	}
	policy.AddObject(contract)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "xyz",
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

	contract := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Contexts: []*Context{{
			Name: "special-invalid-context-keys",
			Allocation: &Allocation{
				Service: "xyz",
				Keys: []string{
					"wowowow {{{{.......",
				},
			},
		}},
	}
	policy.AddObject(contract)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "xyz",
	}
	policy.AddObject(dependency)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Owner doesn't exist")
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

	contract := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Contexts: []*Context{{
			Name: "context",
			Allocation: &Allocation{
				Service: "xyz",
			},
		}},
	}
	policy.AddObject(contract)

	dependency1 := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new_1",
		},
		UserID:   "7",
		Contract: "xyz",
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
		UserID:   "7",
		Contract: "xyz",
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
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
			},
		},
	}
	policy.AddObject(service)

	contract := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Contexts: []*Context{{
			Name: "context",
			Allocation: &Allocation{
				Service: "xyz",
			},
		}},
	}
	policy.AddObject(contract)

	dependency1 := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new_1",
		},
		UserID:   "7",
		Contract: "xyz",
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
		UserID:   "7",
		Contract: "xyz",
		Labels: map[string]string{
			"deplabel": "2",
		},
	}
	policy.AddObject(dependency2)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Conflicting discovery parameters")
}

func TestPolicyResolverInvalidCodeParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	serviceA := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceA",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:     "component",
				Contract: "contractB",
			},
		},
	}
	policy.AddObject(serviceA)

	serviceB := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceB",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component",
				Code: &Code{
					Type: "aptomi/code/unittests",
					Params: util.NestedParameterMap{
						"address": "{{ ..... invalid",
					},
				},
			},
		},
	}
	policy.AddObject(serviceB)

	contractA := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractA",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceA",
			},
		}},
	}
	contractB := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractB",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceB",
			},
		}},
	}
	policy.AddObject(contractA)
	policy.AddObject(contractB)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	}
	policy.AddObject(dependency)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Error when processing code params")
}

func TestPolicyResolverInvalidDiscoveryParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	serviceA := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceA",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:     "component",
				Contract: "contractB",
			},
		},
	}
	policy.AddObject(serviceA)

	serviceB := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceB",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component",
				Discovery: util.NestedParameterMap{
					"address": "{{ .... invalid",
				},
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
			},
		},
	}
	policy.AddObject(serviceB)

	contractA := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractA",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceA",
			},
		}},
	}
	contractB := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractB",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceB",
			},
		}},
	}
	policy.AddObject(contractA)
	policy.AddObject(contractB)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	}
	policy.AddObject(dependency)

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Error when processing discovery params")
}

func TestPolicyResolverServiceLoop(t *testing.T) {
	policy := loadUnitTestsPolicy()

	serviceA := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceA",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:     "component",
				Contract: "contractB",
			},
		},
	}
	policy.AddObject(serviceA)

	serviceB := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceB",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:     "component",
				Contract: "contractC",
			},
		},
	}
	policy.AddObject(serviceB)

	serviceC := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceC",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name:     "component",
				Contract: "contractA",
			},
		},
	}
	policy.AddObject(serviceC)

	contractA := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractA",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceA",
			},
		}},
	}
	contractB := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractB",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceB",
			},
		}},
	}
	contractC := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractC",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceC",
			},
		}},
	}
	policy.AddObject(contractA)
	policy.AddObject(contractB)
	policy.AddObject(contractC)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	}
	policy.AddObject(dependency)

	// policy with cycle should not be resolved successfully
	resolvePolicy(t, policy, ResError, "cycle detected")
}

func TestPolicyResolverComponentLoop(t *testing.T) {
	policy := loadUnitTestsPolicy()

	serviceA := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceA",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component1",
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
				Dependencies: []string{
					"component2",
				},
			},
			{
				Name: "component2",
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
				Dependencies: []string{
					"component3",
				},
			},
			{
				Name: "component3",
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
				Dependencies: []string{
					"component1",
				},
			},
		},
	}
	policy.AddObject(serviceA)

	contractA := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractA",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceA",
			},
		}},
	}
	policy.AddObject(contractA)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	}
	policy.AddObject(dependency)

	// policy with component cycle should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Component cycle detected")
}

func TestPolicyResolverUnknownComponentType(t *testing.T) {
	policy := loadUnitTestsPolicy()

	serviceA := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "serviceA",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component-unknown",
			},
			{
				Name: "component1",
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
				Dependencies: []string{
					"component2",
				},
			},
			{
				Name: "component2",
				Code: &Code{
					Type: "aptomi/code/unittests",
				},
			},
		},
	}
	policy.AddObject(serviceA)

	contractA := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contractA",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "serviceA",
			},
		}},
	}
	policy.AddObject(contractA)

	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	}
	policy.AddObject(dependency)

	// policy with cycle should not be resolved successfully
	resolvePolicy(t, policy, ResSuccess, "")
}
