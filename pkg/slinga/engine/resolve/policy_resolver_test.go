package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func TestPolicyResolverAndResolvedData(t *testing.T) {
	policy, resolution := loadPolicyAndResolve(t)

	// Check that policy resolution finished correctly
	assert.Equal(t, 12, len(resolution.ComponentInstanceMap), "Policy resolution data should have correct number of entries")

	// Resolution for test context
	kafkaTest := getInstanceByParams(t, "main", "cluster-us-east", "kafka", "test", []string{"platform_services"}, "component2", policy, resolution)
	assert.Equal(t, 1, len(kafkaTest.DependencyKeys), "One dependency should be resolved with access to test, but found %v", kafkaTest.DependencyKeys)
	assert.NotEmpty(t, resolution.DependencyInstanceMap["main:dependency:dep_id_1"], "Alice should have access to test")

	// Resolution for prod context
	kafkaProd := getInstanceByParams(t, "main", "cluster-us-east", "kafka", "prod-low", []string{"team-platform_services", "true"}, "component2", policy, resolution)
	assert.Equal(t, 1, len(kafkaProd.DependencyKeys), "One dependency should be resolved with access to prod, but found %v", kafkaProd.DependencyKeys)
	assert.NotEmpty(t, resolution.DependencyInstanceMap["main:dependency:dep_id_2"], "Bob should have access to prod")
}

func TestPolicyResolverAndUnresolvedData(t *testing.T) {
	_, resolution := loadPolicyAndResolve(t)

	// Dave dependency on kafka should not be resolved
	assert.Empty(t, resolution.DependencyInstanceMap["main:dependency:dep_id_4"], "Partial matching is broken. User has access to kafka, but not to zookeeper that kafka depends on. This should not be resolved successfully")
}

func TestPolicyResolverLabelProcessing(t *testing.T) {
	_, resolution := loadPolicyAndResolve(t)

	// Check labels for Bob's dependency
	serviceInstance := getInstanceByDependencyKey(t, "main:dependency:dep_id_2", resolution)
	labels := serviceInstance.CalculatedLabels.Labels
	assert.Equal(t, "yes", labels["important"], "Label 'important=yes' should be carried from dependency all the way through the policy")
	assert.Equal(t, "true", labels["prod-low-ctx"], "Label 'prod-low-ctx=true' should be added on context match")
	assert.Equal(t, "", labels["some-label-to-be-removed"], "Label 'some-label-to-be-removed' should be removed on context match")
	assert.Equal(t, "true", labels["prod-low-alloc"], "Label 'prod-low-alloc=true' should be added on allocation match")
}

func TestPolicyResolverCodeAndDiscoveryParamsEval(t *testing.T) {
	policy, resolution := loadPolicyAndResolve(t)

	kafkaTest := getInstanceByParams(t, "main", "cluster-us-east", "kafka", "test", []string{"platform_services"}, "component2", policy, resolution)

	// Check that code parameters evaluate correctly
	assert.Equal(t, strings.Join(
		[]string{"cluster-us-west", "main", "zookeeper", "test", "platform-services", "component2"}, "-",
	), kafkaTest.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")

	// Check that discovery parameters evaluate correctly
	assert.Equal(t, strings.Join(
		[]string{"kafka", "cluster-us-east", "main", "kafka", "test", "platform-services", "component2", "url"}, "-",
	), kafkaTest.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

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
	assert.Empty(t, resolution.DependencyInstanceMap[dependency.GetKey()], "Dependency should not be resolved")
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
	assert.Empty(t, resolution.DependencyInstanceMap[dependency.GetKey()], "Dependency should not be resolved")
}

func TestPolicyResolverInvalidContextCriteria(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Owner: "1",
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "xyz",
	})

	// policy with invalid context should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Unable to compile expression")
}

func TestPolicyResolverInvalidContextKeys(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
		Owner: "1",
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "xyz",
	})

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Error while resolving allocation keys")
}

func TestPolicyResolverInvalidServiceWithoutOwner(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "xyz",
		},
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "xyz",
	})

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Owner doesn't exist")
}

func TestPolicyResolverInvalidRuleCriteria(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "special-invalid-rule-require-all",
		},
		Criteria: &Criteria{
			RequireAll: []string{"specialname + '123')((("},
		},
		Actions: &RuleActions{},
	})

	// policy with invalid rule should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Unable to compile expression")
}

func TestPolicyResolverConflictingCodeParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"service.Name == 'xyz'"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

	policy.AddObject(&Dependency{
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
	})

	policy.AddObject(&Dependency{
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
	})

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Conflicting code parameters")
}

func TestPolicyResolverConflictingDiscoveryParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"service.Name == 'xyz'"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

	policy.AddObject(&Dependency{
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
	})

	policy.AddObject(&Dependency{
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
	})

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Conflicting discovery parameters")
}

func TestPolicyResolverInvalidCodeParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"in(service.Name, 'serviceA', 'serviceB')"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	})

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Error when processing code params")
}

func TestPolicyResolverInvalidDiscoveryParams(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})
	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"in(service.Name, 'serviceA', 'serviceB')"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	})

	// policy with invalid context allocation keys should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Error when processing discovery params")
}

func TestPolicyResolverServiceLoop(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})
	policy.AddObject(&Contract{
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
	})
	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"in(service.Name, 'serviceA', 'serviceB', 'serviceC')"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

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

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"service.Name == 'serviceA'"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "7",
		Contract: "contractA",
	})

	// policy with component cycle should not be resolved successfully
	resolvePolicy(t, policy, ResError, "Component cycle detected")
}

func TestPolicyResolverUnknownComponentType(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
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
	})

	policy.AddObject(&Contract{
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
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 1,
		Criteria: &Criteria{
			RequireAll: []string{"service.Name == 'serviceA'"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

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

	// unknown component type should not result in critical error
	resolution := resolvePolicy(t, policy, ResSuccess, "")

	// check that both dependencies got resolved
	assert.NotEmpty(t, resolution.DependencyInstanceMap[dependency.GetKey()], "Dependency should be resolved")
}

func TestPolicyResolverRulesForTwoClusters(t *testing.T) {
	policy := loadUnitTestsPolicy()

	policy.AddObject(&Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      "service",
		},
		Owner: "1",
		Components: []*ServiceComponent{
			{
				Name: "component",
				Code: &Code{
					Type: "aptomi/code/unittests",
					Params: util.NestedParameterMap{
						LabelCluster: "{{ .Labels.cluster }}",
					},
				},
			},
		},
	})

	policy.AddObject(&Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
			Namespace: "main",
			Name:      "contract",
		},
		Contexts: []*Context{{
			Name: "context",
			Criteria: &Criteria{
				RequireAny: []string{"true"},
			},
			Allocation: &Allocation{
				Service: "service",
			},
		}},
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule1",
		},
		Weight: 100,
		Criteria: &Criteria{
			RequireAll: []string{"label1 == 'value1'"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-east")),
		},
	})

	policy.AddObject(&Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule2",
		},
		Weight: 200,
		Criteria: &Criteria{
			RequireAll: []string{"label2 == 'value2'"},
		},
		Actions: &RuleActions{
			Dependency:   DependencyAction("allow"),
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(LabelCluster, "cluster-us-west")),
		},
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_1",
		},
		UserID: "7",
		Labels: map[string]string{
			"label1": "value1",
		},
		Contract: "contract",
	})

	policy.AddObject(&Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dep_2",
		},
		UserID: "7",
		Labels: map[string]string{
			"label2": "value2",
		},
		Contract: "contract",
	})

	// unknown component type should not result in critical error
	resolution := resolvePolicy(t, policy, ResSuccess, "")

	// check that both dependencies got resolved
	instance1 := getInstanceByDependencyKey(t, "main:dependency:dep_1", resolution)
	instance2 := getInstanceByDependencyKey(t, "main:dependency:dep_2", resolution)
	assert.Equal(t, "cluster-us-east", instance1.CalculatedLabels.Labels[LabelCluster], "Cluster should be set correctly via rules")
	assert.Equal(t, "cluster-us-west", instance2.CalculatedLabels.Labels[LabelCluster], "Cluster should be set correctly via rules")
}
