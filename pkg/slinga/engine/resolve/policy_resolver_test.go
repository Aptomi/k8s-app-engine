package resolve

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/builder"
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
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	user := &lang.User{ID: "non-existing-user-123456789"}
	dependency := b.AddDependency(user, contract)

	// dependency declared by non-existing consumer should not trigger a critical error
	resolution := resolvePolicyNew(t, b, ResSuccess, "non-existing user")
	assert.Empty(t, resolution.DependencyInstanceMap[dependency.GetKey()], "Dependency should not be resolved")
}

func TestPolicyResolverDependencyWithNonExistingContract(t *testing.T) {
	b := builder.NewPolicyBuilder()
	user := b.AddUser()
	contract := &lang.Contract{Metadata: lang.Metadata{Name: "non-existing-contract-123456789"}}
	dependency := b.AddDependency(user, contract)

	// dependency referring to non-existing contract should not trigger a critical error
	resolution := resolvePolicyNew(t, b, ResSuccess, "non-existing contract")
	assert.Empty(t, resolution.DependencyInstanceMap[dependency.GetKey()], "Dependency should not be resolved")
}

func TestPolicyResolverInvalidContextCriteria(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	criteria := b.Criteria("true", "specialname + '123')(((", "false")
	contract := b.AddContract(service, criteria, nil)
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid context criteria should result in an error
	resolvePolicyNew(t, b, ResError, "Unable to compile expression")
}

func TestPolicyResolverInvalidContextKeys(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue(), b.AllocationKeys("w {{..."))
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid context allocation keys should result in an error
	resolvePolicyNew(t, b, ResError, "Error while resolving allocation keys")
}

func TestPolicyResolverInvalidServiceWithoutOwner(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(nil)
	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid service (without owner) should result in an error
	resolvePolicyNew(t, b, ResError, "Owner doesn't exist for service")
}

func TestPolicyResolverInvalidRuleCriteria(t *testing.T) {
	b := builder.NewPolicyBuilder()
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	b.AddDependency(b.AddUser(), contract)
	b.AddRule(b.Criteria("specialname + '123')(((", "true", "false"), nil)

	// policy resolution with invalid rule should result in an error
	resolvePolicyNew(t, b, ResError, "Unable to compile expression")
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
	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["deplabel"] = "1"

	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["deplabel"] = "2"

	// policy resolution with conflicting code parameters should result in an error
	resolvePolicyNew(t, b, ResError, "Conflicting code parameters")
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

	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependencies which feed conflicting labels into a given component
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["deplabel"] = "1"

	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["deplabel"] = "2"

	// policy resolution with conflicting discovery parameters should result in an error
	resolvePolicyNew(t, b, ResError, "Conflicting discovery parameters")
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
	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid code parameters should result in an error
	resolvePolicyNew(t, b, ResError, "Error when processing code params")
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
	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with invalid discovery parameters should result in an error
	resolvePolicyNew(t, b, ResError, "Error when processing discovery params")
}

func TestPolicyResolverServiceLoop(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create 3 services
	service1 := b.AddService(b.AddUser())
	contract1 := b.AddContract(service1, b.CriteriaTrue(), nil)
	service2 := b.AddService(b.AddUser())
	contract2 := b.AddContract(service2, b.CriteriaTrue(), nil)
	service3 := b.AddService(b.AddUser())
	contract3 := b.AddContract(service3, b.CriteriaTrue(), nil)

	// create service-level cycle
	b.AddServiceComponent(service1, b.ContractComponent(contract2))
	b.AddServiceComponent(service2, b.ContractComponent(contract3))
	b.AddServiceComponent(service3, b.ContractComponent(contract1))

	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract1)

	// policy resolution with service dependency cycle should should result in an error
	resolvePolicyNew(t, b, ResError, "service cycle detected")
}

func TestPolicyResolverComponentLoop(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create 3 services
	service := b.AddService(b.AddUser())
	contract := b.AddContract(service, b.CriteriaTrue(), nil)

	// create component cycle
	component1 := b.CodeComponent(nil, nil)
	component2 := b.CodeComponent(nil, nil)
	component3 := b.CodeComponent(nil, nil)
	b.AddComponentDependency(component1, component2)
	b.AddComponentDependency(component2, component3)
	b.AddComponentDependency(component3, component1)
	b.AddServiceComponent(service, component1)
	b.AddServiceComponent(service, component2)
	b.AddServiceComponent(service, component3)

	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	b.AddDependency(b.AddUser(), contract)

	// policy resolution with service dependency cycle should should result in an error
	resolvePolicyNew(t, b, ResError, "Component cycle detected")
}

func TestPolicyResolverUnknownComponentType(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a with 3 components, first component is not code and not contract (engine should just skip it)
	service := b.AddService(b.AddUser())
	component1 := b.UnknownComponent()
	component2 := b.CodeComponent(nil, nil)
	component3 := b.CodeComponent(nil, nil)
	b.AddComponentDependency(component1, component2)
	b.AddComponentDependency(component2, component3)

	b.AddServiceComponent(service, component1)
	b.AddServiceComponent(service, component2)
	b.AddServiceComponent(service, component3)

	contract := b.AddContract(service, b.CriteriaTrue(), nil)
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster.Name)))

	// add dependency
	dependency := b.AddDependency(b.AddUser(), contract)

	// unknown component type should not result in critical error
	resolution := resolvePolicyNew(t, b, ResSuccess, "Skipping unknown component")

	// check that both dependencies got resolved
	assert.NotEmpty(t, resolution.DependencyInstanceMap[dependency.GetKey()], "Dependency should be resolved")
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
	contract := b.AddContract(service, b.CriteriaTrue(), nil)

	// add rules, which say to deploy to different clusters based on label value
	cluster1 := b.AddCluster()
	cluster2 := b.AddCluster()
	b.AddRule(b.Criteria("label1 == 'value1'", "true", "false"), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster1.Name)))
	b.AddRule(b.Criteria("label2 == 'value2'", "true", "false"), b.RuleActions(lang.Allow, lang.Allow, lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, cluster2.Name)))

	// add dependencies
	d1 := b.AddDependency(b.AddUser(), contract)
	d1.Labels["label1"] = "value1"
	d2 := b.AddDependency(b.AddUser(), contract)
	d2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicyNew(t, b, ResSuccess, "Successfully resolved")

	// check that both dependencies got resolved and got placed in different clusters
	instance1 := getInstanceByDependencyKey(t, d1.GetKey(), resolution)
	instance2 := getInstanceByDependencyKey(t, d2.GetKey(), resolution)
	assert.Equal(t, cluster1.Name, instance1.CalculatedLabels.Labels[lang.LabelCluster], "Cluster should be set correctly via rules")
	assert.Equal(t, cluster2.Name, instance2.CalculatedLabels.Labels[lang.LabelCluster], "Cluster should be set correctly via rules")
}
