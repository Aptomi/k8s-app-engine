package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyResolverSimple(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service with two contexts within a contract
	service := b.AddService()
	component := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	contract := b.AddContractMultipleContexts(service,
		b.Criteria("label1 == 'value1'", "true", "false"),
		b.Criteria("label2 == 'value2'", "true", "false"),
	)

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependency (should be resolved to the first context)
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["label1"] = "value1"

	// add dependency (should be resolved to the second context)
	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d1), "Dependency should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d1)].Resolved, "Dependency should be successfully resolved")
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d2), "Dependency should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d2)].Resolved, "Dependency should be successfully resolved")

	// check instance 1
	instance1 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[0], nil, service, component, resolution)
	assert.Equal(t, 1, len(instance1.DependencyKeys), "Instance should be referenced by one dependency")

	// check instance 2
	instance2 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[1], nil, service, component, resolution)
	assert.Equal(t, 1, len(instance2.DependencyKeys), "Instance should be referenced by one dependency")
}

func TestPolicyResolverComponentWithCriteria(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service with conditional components
	service := b.AddService()
	contract := b.AddContract(service, b.CriteriaTrue())

	component1 := b.CodeComponent(nil, nil)
	component1.Criteria = b.CriteriaTrue()
	b.AddServiceComponent(service, component1)

	component2 := b.CodeComponent(nil, nil)
	component2.Criteria = &lang.Criteria{RequireAll: []string{"param2 == 'value2'"}}
	b.AddServiceComponent(service, component2)

	component3 := b.CodeComponent(nil, nil)
	component3.Criteria = &lang.Criteria{RequireAll: []string{"param3 == 'value3'"}}
	b.AddServiceComponent(service, component3)

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependency (should be resolved to use components 1 and 2, but not 3)
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["param2"] = "value2"

	// add dependency (should be resolved to use components 1 and 3, but not 2)
	d2 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["param3"] = "value3"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d1), "Dependency should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d1)].Resolved, "Dependency should be successfully resolved")
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d2), "Dependency should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d2)].Resolved, "Dependency should be successfully resolved")

	// check component 1
	instance1 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[0], nil, service, component1, resolution)
	assert.Equal(t, 2, len(instance1.DependencyKeys), "Component 1 instance should be used by both dependencies")

	// check component 2
	instance2 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[0], nil, service, component2, resolution)
	assert.Equal(t, 1, len(instance2.DependencyKeys), "Component 2 instance should be used by only one dependency")

	// check component 3
	instance3 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[0], nil, service, component3, resolution)
	assert.Equal(t, 1, len(instance3.DependencyKeys), "Component 3 instance should be used by only one dependency")
}

func TestPolicyResolverMultipleNS(t *testing.T) {
	b := builder.NewPolicyBuilder()

	cluster := b.AddCluster()

	// create objects in ns1
	b.SwitchNamespace("ns1")
	service1 := b.AddService()
	b.AddServiceComponent(service1, b.CodeComponent(nil, nil))
	contract1 := b.AddContract(service1, b.CriteriaTrue())
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// create objects in ns2
	b.SwitchNamespace("ns2")
	service2 := b.AddService()
	b.AddServiceComponent(service2, b.CodeComponent(nil, nil))
	contract2 := b.AddContract(service2, b.CriteriaTrue())
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// ns1/service1 -> depends on -> ns2/contract2
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// create dependency in ns3 on ns1/contract1 (it's created on behalf of domain admin, who can for sure consume services from all namespaces)
	b.SwitchNamespace("ns3")
	d := b.AddDependency(b.AddUserDomainAdmin(), contract1)

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d), "Dependency should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d)].Resolved, "Dependency should be successfully resolved")
}

func TestPolicyResolverPartialMatching(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service, which depends on another service
	service1 := b.AddService()
	contract1 := b.AddContract(service1, b.Criteria("label1 == 'value1'", "true", "false"))
	service2 := b.AddService()
	contract2 := b.AddContract(service2, b.Criteria("label2 == 'value2'", "true", "false"))
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependency with full labels (should be resolved successfully)
	d1 := b.AddDependency(b.AddUser(), contract1)
	d1.Labels["label1"] = "value1"
	d1.Labels["label2"] = "value2"

	// add dependency with partial labels (should not resolved)
	d2 := b.AddDependency(b.AddUser(), contract1)
	d2.Labels["label1"] = "value1"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResSomeDependenciesFailed, "unable to find matching context")

	// check that only first dependency got resolved
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d1), "Dependency with full set of labels should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d1)].Resolved, "Dependency with full set of labels should be successfully resolved")

	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d2), "Dependency with partial set of labels should be present in policy resolution")
	assert.False(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d2)].Resolved, "Dependency with partial set of labels should not be successfully resolved")
}

func TestPolicyResolverCalculatedLabels(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// first contract adds a label 'labelExtra1'
	service1 := b.AddService()
	contract1 := b.AddContract(service1, b.Criteria("label1 == 'value1'", "true", "false"))
	contract1.ChangeLabels = lang.NewLabelOperationsSetSingleLabel("labelExtra1", "labelValue1")

	// second contract adds a label 'labelExtra2' and removes 'label3'
	service2 := b.AddService()
	contract2 := b.AddContract(service2, b.Criteria("label2 == 'value2'", "true", "false"))
	contract2.ChangeLabels = lang.NewLabelOperations(
		map[string]string{"labelExtra2": "labelValue2"},
		map[string]string{"label3": ""},
	)
	contract2.Contexts[0].ChangeLabels = lang.NewLabelOperationsSetSingleLabel("labelExtra3", "labelValue3")
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependency with labels 'label1', 'label2' and 'label3'
	dependency := b.AddDependency(b.AddUser(), contract1)
	dependency.Labels["label1"] = "value1"
	dependency.Labels["label2"] = "value2"
	dependency.Labels["label3"] = "value3"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")

	// check that dependency got resolved
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(dependency), "Dependency should be present in policy resolution")
	assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(dependency)].Resolved, "Dependency should be successfully resolved")

	// check labels for the end service (service2/contract2)
	serviceInstance := getInstanceByParams(t, cluster, "k8ns", contract2, contract2.Contexts[0], nil, service2, nil, resolution)
	labels := serviceInstance.CalculatedLabels.Labels

	assert.Equal(t, "value1", labels["label1"], "Label 'label1=value1' should be carried from dependency all the way through the policy")
	assert.Equal(t, "value2", labels["label2"], "Label 'label2=value2' should be carried from dependency all the way through the policy")
	assert.NotContains(t, labels, "label3", "Label 'label3' should be removed")

	assert.Equal(t, "labelValue1", labels["labelExtra1"], "Label 'labelExtra1' should be added on contract match")
	assert.Equal(t, "labelValue2", labels["labelExtra2"], "Label 'labelExtra2' should be added on contract match")
	assert.Equal(t, "labelValue3", labels["labelExtra3"], "Label 'labelExtra3' should be added on context match")

	assert.Equal(t, cluster.Name, labels[lang.LabelTarget], "Label 'cluster' should be set")
}

func TestPolicyResolverCodeAndDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service with 2 components and multiple parameters
	service := b.AddService()
	component1 := b.CodeComponent(
		nil,
		util.NestedParameterMap{"url": "component1-{{ .Discovery.Instance }}"},
	)
	component2 := b.CodeComponent(
		util.NestedParameterMap{
			"debug":   "{{ .Labels.target }}",
			"address": fmt.Sprintf("{{ .Discovery.%s.url }}", component1.Name),
			"nested": util.NestedParameterMap{
				"param": util.NestedParameterMap{
					"name1":    "value1",
					"name2":    "123456789",
					"debug":    "{{ .Labels.target }}",
					"nameBool": true,
					"nameInt":  5,
				},
			},
		},
		util.NestedParameterMap{"url": "component2-{{ .Discovery.Instance }}"},
	)
	b.AddServiceComponent(service, component1)
	b.AddServiceComponent(service, component2)

	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	b.AddDependency(b.AddUser(), contract)

	// policy should be resolved successfully
	resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")

	// check discovery parameters of component 1
	instance1 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[0], nil, service, component1, resolution)
	assert.Regexp(t, "^component1-(.+)$", instance1.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// check discovery parameters of component 2
	instance2 := getInstanceByParams(t, cluster, "k8ns", contract, contract.Contexts[0], nil, service, component2, resolution)
	assert.Regexp(t, "^component2-(.+)$", instance2.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// check code parameters of component 2
	assert.Equal(t, cluster.Name, instance2.CalculatedCodeParams["debug"], "Code parameter should be calculated correctly")
	assert.Equal(t, instance1.CalculatedDiscovery["url"], instance2.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")
	assert.Equal(t, "value1", instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["name1"], "Code parameter should be calculated correctly")
	assert.Equal(t, "123456789", instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["name2"], "Code parameter should be calculated correctly")
	assert.Equal(t, cluster.Name, instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["debug"], "Code parameter should be calculated correctly")
	assert.Equal(t, true, instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["nameBool"], "Code parameter should be calculated correctly (bool)")
	assert.Equal(t, 5, instance2.CalculatedCodeParams.GetNestedMap("nested").GetNestedMap("param")["nameInt"], "Code parameter should be calculated correctly (int)")
}

func TestPolicyResolverDependencyWithNonExistingUser(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService()
	contract := b.AddContract(service, b.CriteriaTrue())
	user := &lang.User{Name: "non-existing-user-123456789"}
	dependency := b.AddDependency(user, contract)

	// dependency declared by non-existing consumer should result in dependency error
	resolution := resolvePolicy(t, b, ResSomeDependenciesFailed, "non-existing user")
	assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(dependency), "Dependency should be present in policy resolution")
	assert.False(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(dependency)].Resolved, "Dependency should not be successfully resolved")
}

func TestPolicyResolverConflictingCodeParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which uses label in its code parameters
	service := b.AddService()
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{"address": "{{ .Labels.deplabel }}"},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["deplabel"] = "1"

	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["deplabel"] = "2"

	// policy resolution with conflicting code parameters should result in an error
	resolvePolicy(t, b, ResSomeDependenciesFailed, "conflicting code parameters")
}

func TestPolicyResolverConflictingDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which uses label in its discovery parameters
	service := b.AddService()
	b.AddServiceComponent(service,
		b.CodeComponent(
			nil,
			util.NestedParameterMap{"address": "{{ .Labels.deplabel }}"},
		),
	)

	contract := b.AddContract(service, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["deplabel"] = "1"

	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["deplabel"] = "2"

	// policy resolution with conflicting discovery parameters should result in an error
	resolvePolicy(t, b, ResSomeDependenciesFailed, "conflicting discovery parameters")
}

func TestPolicyResolverServiceLoop(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create 3 services
	service1 := b.AddService()
	contract1 := b.AddContract(service1, b.CriteriaTrue())
	service2 := b.AddService()
	contract2 := b.AddContract(service2, b.CriteriaTrue())
	service3 := b.AddService()
	contract3 := b.AddContract(service3, b.CriteriaTrue())

	// create service-level cycle
	b.AddServiceComponent(service1, b.ContractComponent(contract2))
	b.AddServiceComponent(service2, b.ContractComponent(contract3))
	b.AddServiceComponent(service3, b.ContractComponent(contract1))

	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract1)

	// policy resolution with service dependency cycle should should result in an error
	resolvePolicy(t, b, ResSomeDependenciesFailed, "service cycle detected")
}

func TestPolicyResolverPickClusterViaRules(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a service which can be deployed to different clusters
	service := b.AddService()
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{"debug": "{{ .Labels.target }}"},
			nil,
		),
	)
	contract := b.AddContract(service, b.CriteriaTrue())

	// add rules, which say to deploy to different clusters based on label value
	cluster1 := b.AddCluster()
	cluster2 := b.AddCluster()
	b.AddRule(b.Criteria("label1 == 'value1'", "true", "false"), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster1.Name)))
	b.AddRule(b.Criteria("label2 == 'value2'", "true", "false"), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster2.Name)))

	// add dependencies
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["label1"] = "value1"
	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")

	// check that both dependencies got resolved and got placed in different clusters
	instance1 := getInstanceByDependencyKey(t, runtime.KeyForStorable(d1), resolution)
	instance2 := getInstanceByDependencyKey(t, runtime.KeyForStorable(d2), resolution)
	assert.Equal(t, cluster1.Name, instance1.CalculatedLabels.Labels[lang.LabelTarget], "Cluster should be set correctly via rules")
	assert.Equal(t, cluster2.Name, instance2.CalculatedLabels.Labels[lang.LabelTarget], "Cluster should be set correctly via rules")
}

func TestPolicyResolverInternalPanic(t *testing.T) {
	b := builder.NewPolicyBuilder()
	b.PanicWhenLoadingUsers()

	// create a service with two contexts within a contract
	service := b.AddService()
	b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	contract := b.AddContract(service, b.CriteriaTrue())

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add multiple dependencies
	for i := 0; i < 10; i++ {
		b.AddDependency(b.AddUser(), contract)
	}

	// policy resolution should result in an error
	resolvePolicy(t, b, ResSomeDependenciesFailed, "panic from mock user loader")
}

func TestPolicyResolverAllocationKeys(t *testing.T) {
	b := builder.NewPolicyBuilder()

	tc := map[string]bool{
		"fixedValue":           true,
		"{{ .User.Name }}":     false,
		"{{ .Dependency.ID }}": false,
	}

	for allocationKey, sameInstance := range tc {
		// create a service and a contract
		service := b.AddService()
		b.AddServiceComponent(service, b.CodeComponent(nil, nil))
		contract := b.AddContract(service, b.CriteriaTrue())
		contract.Contexts[0].Allocation.Keys = b.AllocationKeys(allocationKey)

		// add rule to set cluster
		cluster := b.AddCluster()
		b.AddRule(b.Criteria(
			"Service.Namespace == 'main' && Dependency.Namespace == 'main'",
			"true",
			"false",
		), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

		// add dependency (should be resolved to the first context)
		d1 := b.AddDependency(b.AddUser(), contract)

		// add dependency (should be resolved to the second context)
		d2 := b.AddDependency(b.AddUser(), contract)

		// policy resolution should be completed successfully
		resolution := resolvePolicy(t, b, ResAllDependenciesResolvedSuccessfully, "Successfully resolved")
		assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d1), "Dependency should be present in policy resolution")
		assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d1)].Resolved, "Dependency should be successfully resolved")
		assert.Contains(t, resolution.GetDependencyInstanceMap(), runtime.KeyForStorable(d2), "Dependency should be present in policy resolution")
		assert.True(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d2)].Resolved, "Dependency should be successfully resolved")

		// make sure dependencies point to different instances
		if sameInstance {
			assert.Equal(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d1)].ComponentInstanceKey, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d2)].ComponentInstanceKey, "Dependencies should point to same instance")
		} else {
			assert.NotEqual(t, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d1)].ComponentInstanceKey, resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(d2)].ComponentInstanceKey, "Dependencies should point to different instances")
		}
	}
}

/*
	Helpers
*/

const (
	ResAllDependenciesResolvedSuccessfully = iota
	ResSomeDependenciesFailed              = iota
)

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder, expectedResult int, expectedLogMessage string) *PolicyResolution {
	t.Helper()
	eventLog := event.NewLog(logrus.DebugLevel, "test-resolve")
	resolver := NewPolicyResolver(builder.Policy(), builder.External(), eventLog)
	result := resolver.ResolveAllDependencies()

	if !assert.Equal(t, expectedResult != ResSomeDependenciesFailed, result.AllDependenciesResolvedSuccessfully(), "All dependencies should be resolved successfully") {
		// print log into stdout and exit
		hook := event.NewHookConsole(logrus.DebugLevel)
		eventLog.Save(hook)
		t.FailNow()
		return nil
	}

	// check for error message
	verifier := event.NewLogVerifier(expectedLogMessage, expectedResult == ResSomeDependenciesFailed)
	resolver.eventLog.Save(verifier)
	if !assert.True(t, verifier.MatchedErrorsCount() > 0, "Event log should have an error message containing words: %s", expectedLogMessage) {
		hook := event.NewHookConsole(logrus.DebugLevel)
		resolver.eventLog.Save(hook)
		t.FailNow()
	}

	return result
}

func getInstanceByDependencyKey(t *testing.T, dependencyID string, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	dResolution := resolution.GetDependencyInstanceMap()[dependencyID]
	if !assert.True(t, dResolution.Resolved, "Dependency %s should be resolved", dependencyID) {
		t.Log(resolution.GetDependencyInstanceMap())
		t.FailNow()
	}
	instance, ok := resolution.ComponentInstanceMap[dResolution.ComponentInstanceKey]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", dResolution.ComponentInstanceKey) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, cluster *lang.Cluster, namespace string, contract *lang.Contract, context *lang.Context, allocationKeysResolved []string, service *lang.Service, component *lang.ServiceComponent, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := NewComponentInstanceKey(cluster, namespace, contract, context, allocationKeysResolved, service, component)
	instance, ok := resolution.ComponentInstanceMap[key.GetKey()]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key.GetKey()) {
		t.FailNow()
	}
	return instance
}
