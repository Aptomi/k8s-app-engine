package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyResolverContract(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service with two contexts within a contract
	service := b.AddService(b.AddUser())
	component := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	contract := b.AddContractMultipleContexts(service,
		b.Criteria("label1 == 'value1'", "true", "false"),
		b.Criteria("label2 == 'value2'", "true", "false"),
	)

	// add rules to allow all dependencies
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency (should be resolved to the first context)
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["label1"] = "value1"

	// add dependency (should be resolved to the second context)
	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResSuccess, "Successfully resolved")
	assert.Contains(t, resolution.DependencyInstanceMap, object.GetKey(d1), "Dependency should be resolved")
	assert.Contains(t, resolution.DependencyInstanceMap, object.GetKey(d2), "Dependency should be resolved")

	// check instance 1
	instance1 := getInstanceByParams(t, cluster, contract, contract.Contexts[0], nil, service, component, resolution)
	assert.Equal(t, 1, len(instance1.DependencyKeys), "Instance should be referenced by one dependency")

	// check instance 2
	instance2 := getInstanceByParams(t, cluster, contract, contract.Contexts[1], nil, service, component, resolution)
	assert.Equal(t, 1, len(instance2.DependencyKeys), "Instance should be referenced by one dependency")
}

func TestPolicyResolverMultipleNS(t *testing.T) {
	b := builder.NewPolicyBuilder()

	cluster := b.AddCluster()

	// create objects in ns1
	b.SwitchNamespace("ns1")
	service1 := b.AddService(b.AddUser())
	b.AddServiceComponent(service1, b.CodeComponent(nil, nil))
	contract1 := b.AddContract(service1, b.CriteriaTrue())
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// create objects in ns2
	b.SwitchNamespace("ns2")
	service2 := b.AddService(b.AddUser())
	b.AddServiceComponent(service2, b.CodeComponent(nil, nil))
	contract2 := b.AddContract(service2, b.CriteriaTrue())
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// ns1/service1 -> depends on -> ns2/contract2
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// create dependency in ns3 on ns1/contract1 (it's created on behalf of domain admin, who can for sure consume services from all namespaces)
	b.SwitchNamespace("ns3")
	d := b.AddDependency(b.AddUserDomainAdmin(), contract1)

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResSuccess, "Successfully resolved")
	assert.Contains(t, resolution.DependencyInstanceMap, object.GetKey(d), "Dependency should be resolved")
}

func TestPolicyResolverPartialMatching(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service, which depends on another service
	service1 := b.AddService(b.AddUser())
	contract1 := b.AddContract(service1, b.Criteria("label1 == 'value1'", "true", "false"))
	service2 := b.AddService(b.AddUser())
	contract2 := b.AddContract(service2, b.Criteria("label2 == 'value2'", "true", "false"))
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// add rules to allow all dependencies
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency with full labels (should be resolved successfully)
	d1 := b.AddDependency(b.AddUser(), contract1)
	d1.Labels["label1"] = "value1"
	d1.Labels["label2"] = "value2"

	// add dependency with partial labels (should not resolved)
	d2 := b.AddDependency(b.AddUser(), contract1)
	d2.Labels["label1"] = "value1"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResSuccess, "Successfully resolved")

	// check that only first dependency got resolved
	assert.Contains(t, resolution.DependencyInstanceMap, object.GetKey(d1), "Dependency with full set of labels should be resolved")
	assert.NotContains(t, resolution.DependencyInstanceMap, object.GetKey(d2), "Dependency with partial labels should not be resolved")
}

func TestPolicyResolverCalculatedLabels(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// first contract adds a label 'labelExtra1'
	service1 := b.AddService(b.AddUser())
	contract1 := b.AddContract(service1, b.Criteria("label1 == 'value1'", "true", "false"))
	contract1.ChangeLabels = lang.NewLabelOperationsSetSingleLabel("labelExtra1", "labelValue1")

	// second contract adds a label 'labelExtra2' and removes 'label3'
	service2 := b.AddService(b.AddUser())
	contract2 := b.AddContract(service2, b.Criteria("label2 == 'value2'", "true", "false"))
	contract2.ChangeLabels = lang.NewLabelOperations(
		map[string]string{"labelExtra2": "labelValue2"},
		map[string]string{"label3": ""},
	)
	contract2.Contexts[0].ChangeLabels = lang.NewLabelOperationsSetSingleLabel("labelExtra3", "labelValue3")
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// add rules to allow all dependencies
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency with labels 'label1', 'label2' and 'label3'
	dependency := b.AddDependency(b.AddUser(), contract1)
	dependency.Labels["label1"] = "value1"
	dependency.Labels["label2"] = "value2"
	dependency.Labels["label3"] = "value3"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResSuccess, "Successfully resolved")

	// check that dependency got resolved
	assert.Contains(t, resolution.DependencyInstanceMap, object.GetKey(dependency), "Dependency should be resolved")

	// check labels for the end service (service2/contract2)
	serviceInstance := getInstanceByParams(t, cluster, contract2, contract2.Contexts[0], nil, service2, nil, resolution)
	labels := serviceInstance.CalculatedLabels.Labels

	assert.Equal(t, "value1", labels["label1"], "Label 'label1=value1' should be carried from dependency all the way through the policy")
	assert.Equal(t, "value2", labels["label2"], "Label 'label2=value2' should be carried from dependency all the way through the policy")
	assert.NotContains(t, labels, "label3", "Label 'label3' should be removed")

	assert.Equal(t, "labelValue1", labels["labelExtra1"], "Label 'labelExtra1' should be added on contract match")
	assert.Equal(t, "labelValue2", labels["labelExtra2"], "Label 'labelExtra2' should be added on contract match")
	assert.Equal(t, "labelValue3", labels["labelExtra3"], "Label 'labelExtra3' should be added on context match")

	assert.Equal(t, cluster.Name, labels[lang.LabelCluster], "Label 'cluster' should be set")
}

func TestPolicyResolverCodeAndDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service with 2 components and multiple parameters
	service := b.AddService(b.AddUser())
	component1 := b.CodeComponent(
		nil,
		util.NestedParameterMap{"url": "component1-{{ .Discovery.instance }}"},
	)
	component2 := b.CodeComponent(
		util.NestedParameterMap{
			"cluster": "{{ .Labels.cluster }}",
			"address": fmt.Sprintf("{{ .Discovery.%s.url }}", component1.Name),
			"nested": util.NestedParameterMap{
				"param": util.NestedParameterMap{
					"name1": "value1",
					"name2": "123456789",
					"name3": "{{ .Labels.cluster }}",
				},
			},
		},
		util.NestedParameterMap{"url": "component2-{{ .Discovery.instance }}"},
	)
	b.AddServiceComponent(service, component1)
	b.AddServiceComponent(service, component2)

	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	b.AddDependency(b.AddUser(), contract)

	// policy should be resolved successfully
	resolution := resolvePolicy(t, b, ResSuccess, "Successfully resolved")

	// check discovery parameters of component 1
	instance1 := getInstanceByParams(t, cluster, contract, contract.Contexts[0], nil, service, component1, resolution)
	assert.Regexp(t, "^component1-(.+)$", instance1.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// check discovery parameters of component 2
	instance2 := getInstanceByParams(t, cluster, contract, contract.Contexts[0], nil, service, component2, resolution)
	assert.Regexp(t, "^component2-(.+)$", instance2.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// check code parameters of component 2
	assert.Equal(t, cluster.Name, instance2.CalculatedCodeParams["cluster"], "Code parameter should be calculated correctly")
	assert.Equal(t, instance1.CalculatedDiscovery["url"], instance2.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")
	assert.Equal(t, "value1", instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["name1"], "Code parameter should be calculated correctly")
	assert.Equal(t, "123456789", instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["name2"], "Code parameter should be calculated correctly")
	assert.Equal(t, cluster.Name, instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["name3"], "Code parameter should be calculated correctly")
}

func TestPolicyResolverDependencyWithNonExistingUser(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue())
	user := &lang.User{ID: "non-existing-user-123456789"}
	dependency := b.AddDependency(user, contract)

	// dependency declared by non-existing consumer should not trigger a critical error
	resolution := resolvePolicy(t, b, ResSuccess, "non-existing user")
	assert.NotContains(t, resolution.DependencyInstanceMap, object.GetKey(dependency), "Dependency should not be resolved")
}

func TestPolicyResolverDependencyWithNonExistingContract(t *testing.T) {
	b := builder.NewPolicyBuilder()
	user := b.AddUser()
	contract := &lang.Contract{Metadata: lang.Metadata{Name: "non-existing-contract-123456789"}}
	dependency := b.AddDependency(user, contract)

	// dependency referring to non-existing contract should not trigger a critical error
	resolution := resolvePolicy(t, b, ResSuccess, "non-existing contract")
	assert.NotContains(t, resolution.DependencyInstanceMap, object.GetKey(dependency), "Dependency should not be resolved")
}

func TestPolicyResolverInvalidContextCriteria(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	criteria := b.Criteria("true", "specialname + '123')(((", "false")
	contract := b.AddContract(service, criteria)
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid context criteria should result in an error
	resolvePolicy(t, b, ResError, "Unable to compile expression")
}

func TestPolicyResolverInvalidContextKeys(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue())
	contract.Contexts[0].Allocation.Keys = b.AllocationKeys("w {{...")
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid context allocation keys should result in an error
	resolvePolicy(t, b, ResError, "Error while resolving allocation keys")
}

func TestPolicyResolverInvalidServiceWithoutOwner(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(nil)
	contract := b.AddContract(service, b.CriteriaTrue())
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid service (without owner) should result in an error
	resolvePolicy(t, b, ResError, "Owner doesn't exist for service")
}

func TestPolicyResolverInvalidRuleCriteria(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue())
	b.AddDependency(b.AddUser(), contract)
	b.AddRule(b.Criteria("specialname + '123')(((", "true", "false"), nil)

	// policy resolution with invalid rule should result in an error
	resolvePolicy(t, b, ResError, "Unable to compile expression")
}

func TestPolicyResolverConflictingCodeParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which uses label in its code parameters
	service := b.AddService(b.AddUser())
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{"address": "{{ .Labels.deplabel }}"},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["deplabel"] = "1"

	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["deplabel"] = "2"

	// policy resolution with conflicting code parameters should result in an error
	resolvePolicy(t, b, ResError, "Conflicting code parameters")
}

func TestPolicyResolverConflictingDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which uses label in its discovery parameters
	service := b.AddService(b.AddUser())
	b.AddServiceComponent(service,
		b.CodeComponent(
			nil,
			util.NestedParameterMap{"address": "{{ .Labels.deplabel }}"},
		),
	)

	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["deplabel"] = "1"

	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["deplabel"] = "2"

	// policy resolution with conflicting discovery parameters should result in an error
	resolvePolicy(t, b, ResError, "Conflicting discovery parameters")
}

func TestPolicyResolverInvalidCodeParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which uses label in its code parameters
	service := b.AddService(b.AddUser())
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{"address": "{{ .Labels..."},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid code parameters should result in an error
	resolvePolicy(t, b, ResError, "Error when processing code params")
}

func TestPolicyResolverInvalidDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which uses label in its code parameters
	service := b.AddService(b.AddUser())
	b.AddServiceComponent(service,
		b.CodeComponent(
			nil,
			util.NestedParameterMap{"address": "{{ .Labels..."},
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid discovery parameters should result in an error
	resolvePolicy(t, b, ResError, "Error when processing discovery params")
}

func TestPolicyResolverServiceLoop(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create 3 services
	service1 := b.AddService(b.AddUser())
	contract1 := b.AddContract(service1, b.CriteriaTrue())
	service2 := b.AddService(b.AddUser())
	contract2 := b.AddContract(service2, b.CriteriaTrue())
	service3 := b.AddService(b.AddUser())
	contract3 := b.AddContract(service3, b.CriteriaTrue())

	// create service-level cycle
	b.AddServiceComponent(service1, b.ContractComponent(contract2))
	b.AddServiceComponent(service2, b.ContractComponent(contract3))
	b.AddServiceComponent(service3, b.ContractComponent(contract1))

	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract1)

	// policy resolution with service dependency cycle should should result in an error
	resolvePolicy(t, b, ResError, "service cycle detected")
}

func TestPolicyResolverComponentLoop(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create 3 services
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue())

	// create component cycle
	component1 := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	component2 := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	component3 := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	b.AddComponentDependency(component1, component2)
	b.AddComponentDependency(component2, component3)
	b.AddComponentDependency(component3, component1)

	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with service dependency cycle should should result in an error
	resolvePolicy(t, b, ResError, "Component cycle detected")
}

func TestPolicyResolverUnknownComponentType(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a with 3 components, first component is not code and not contract (engine should just skip it)
	service := b.AddService(b.AddUser())
	component1 := b.AddServiceComponent(service, b.UnknownComponent())
	component2 := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	component3 := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	b.AddComponentDependency(component1, component2)
	b.AddComponentDependency(component2, component3)

	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	dependency := b.AddDependency(b.AddUser(), contract)

	// unknown component type should not result in critical error
	resolution := resolvePolicy(t, b, ResSuccess, "Skipping unknown component")

	// check that both dependencies got resolved
	assert.Contains(t, resolution.DependencyInstanceMap, object.GetKey(dependency), "Dependency should be resolved")
}

func TestPolicyResolverPickClusterViaRules(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which can be deployed to different clusters
	service := b.AddService(b.AddUser())
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{lang.LabelCluster: "{{ .Labels.cluster }}"},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())

	// add rules, which say to deploy to different clusters based on label value
	cluster1 := b.AddCluster()
	cluster2 := b.AddCluster()
	b.AddRule(b.Criteria("label1 == 'value1'", "true", "false"), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster1.Name)))
	b.AddRule(b.Criteria("label2 == 'value2'", "true", "false"), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster2.Name)))

	// add dependencies
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["label1"] = "value1"
	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResSuccess, "Successfully resolved")

	// check that both dependencies got resolved and got placed in different clusters
	instance1 := getInstanceByDependencyKey(t, object.GetKey(d1), resolution)
	instance2 := getInstanceByDependencyKey(t, object.GetKey(d2), resolution)
	assert.Equal(t, cluster1.Name, instance1.CalculatedLabels.Labels[lang.LabelCluster], "Cluster should be set correctly via rules")
	assert.Equal(t, cluster2.Name, instance2.CalculatedLabels.Labels[lang.LabelCluster], "Cluster should be set correctly via rules")
}

/*
	Helpers
*/

const (
	ResSuccess = iota
	ResError   = iota
)

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder, expectedResult int, expectedLogMessage string) *PolicyResolution {
	t.Helper()
	resolver := NewPolicyResolver(builder.Policy(), builder.External())
	result, eventLog, err := resolver.ResolveAllDependencies()

	if !assert.Equal(t, expectedResult != ResError, err == nil, "Policy resolution status (success vs. error)") {
		// print log into stdout and exit
		hook := &event.HookStdout{}
		eventLog.Save(hook)
		t.FailNow()
		return nil
	}

	// check for error message
	verifier := event.NewUnitTestLogVerifier(expectedLogMessage, expectedResult == ResError)
	resolver.eventLog.Save(verifier)
	if !assert.True(t, verifier.MatchedErrorsCount() > 0, "Event log should have an error message containing words: "+expectedLogMessage) {
		hook := &event.HookStdout{}
		resolver.eventLog.Save(hook)
		t.FailNow()
	}

	return result
}

func getInstanceByDependencyKey(t *testing.T, dependencyID string, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := resolution.DependencyInstanceMap[dependencyID]
	if !assert.NotZero(t, len(key), "Dependency %s should be resolved", dependencyID) {
		t.Log(resolution.DependencyInstanceMap)
		t.FailNow()
	}
	instance, ok := resolution.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, cluster *lang.Cluster, contract *lang.Contract, context *lang.Context, allocationKeysResolved []string, service *lang.Service, component *lang.ServiceComponent, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := NewComponentInstanceKey(cluster, contract, context, allocationKeysResolved, service, component)
	instance, ok := resolution.ComponentInstanceMap[key.GetKey()]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key.GetKey()) {
		t.FailNow()
	}
	return instance
}
