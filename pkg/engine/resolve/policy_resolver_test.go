// nolint: goconst
package resolve

import (
	"fmt"
	"testing"

	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPolicyResolverSimple(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle with two contexts within a service
	bundle := b.AddBundle()
	component := b.AddBundleComponent(bundle, b.CodeComponent(nil, nil))
	service := b.AddServiceMultipleContexts(bundle,
		b.Criteria("label1 == 'value1'", "true", "false"),
		b.Criteria("label2 == 'value2'", "true", "false"),
	)

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claim (should be resolved to the first context)
	c1 := b.AddClaim(b.AddUser(), service)
	c1.Labels["label1"] = "value1"

	// add claim (should be resolved to the second context)
	c2 := b.AddClaim(b.AddUser(), service)
	c2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, []verifyClaim{
		{claim: c1, resolved: true},
		{claim: c2, resolved: true},
	})

	// check instance 1
	instance1 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[0], nil, bundle, component, resolution)
	assert.Equal(t, 1, len(instance1.ClaimKeys), "Instance should be referenced by one claim")

	// check instance 2
	instance2 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[1], nil, bundle, component, resolution)
	assert.Equal(t, 1, len(instance2.ClaimKeys), "Instance should be referenced by one claim")
}

func TestPolicyResolverComponentWithCriteria(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle with conditional components
	bundle := b.AddBundle()
	service := b.AddService(bundle, b.CriteriaTrue())

	component1 := b.CodeComponent(nil, nil)
	component1.Criteria = b.CriteriaTrue()
	b.AddBundleComponent(bundle, component1)

	component2 := b.CodeComponent(nil, nil)
	component2.Criteria = &lang.Criteria{RequireAll: []string{"param2 == 'value2'"}}
	b.AddBundleComponent(bundle, component2)

	component3 := b.CodeComponent(nil, nil)
	component3.Criteria = &lang.Criteria{RequireAll: []string{"param3 == 'value3'"}}
	b.AddBundleComponent(bundle, component3)

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claim (should be resolved to use components 1 and 2, but not 3)
	c1 := b.AddClaim(b.AddUser(), service)
	c1.Labels["param2"] = "value2"

	// add claim (should be resolved to use components 1 and 3, but not 2)
	c2 := b.AddClaim(b.AddUser(), service)
	c1.Labels["param3"] = "value3"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, []verifyClaim{
		{claim: c1, resolved: true},
		{claim: c2, resolved: true},
	})

	// check component 1
	instance1 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[0], nil, bundle, component1, resolution)
	assert.Equal(t, 2, len(instance1.ClaimKeys), "Component 1 instance should be used by both claims")

	// check component 2
	instance2 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[0], nil, bundle, component2, resolution)
	assert.Equal(t, 1, len(instance2.ClaimKeys), "Component 2 instance should be used by only one claim")

	// check component 3
	instance3 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[0], nil, bundle, component3, resolution)
	assert.Equal(t, 1, len(instance3.ClaimKeys), "Component 3 instance should be used by only one claim")
}

func TestPolicyResolverMultipleNS(t *testing.T) {
	b := builder.NewPolicyBuilder()

	cluster := b.AddCluster()

	// create objects in ns1
	b.SwitchNamespace("ns1")
	bundle1 := b.AddBundle()
	b.AddBundleComponent(bundle1, b.CodeComponent(nil, nil))
	service1 := b.AddService(bundle1, b.CriteriaTrue())
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// create objects in ns2
	b.SwitchNamespace("ns2")
	bundle2 := b.AddBundle()
	b.AddBundleComponent(bundle2, b.CodeComponent(nil, nil))
	service2 := b.AddService(bundle2, b.CriteriaTrue())
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// ns1/bundle1 -> depends on -> ns2/service2
	b.AddBundleComponent(bundle1, b.ServiceComponent(service2))

	// create claim in ns3 on ns1/service1 (it's created on behalf of domain admin, who can for sure consume bundles from all namespaces)
	b.SwitchNamespace("ns3")
	claim := b.AddClaim(b.AddUserDomainAdmin(), service1)

	// policy resolution should be completed successfully
	resolvePolicy(t, b, []verifyClaim{
		{claim: claim, resolved: true},
	})
}

func TestPolicyResolverPartialMatching(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle, which depends on another bundle
	bundle1 := b.AddBundle()
	service1 := b.AddService(bundle1, b.Criteria("label1 == 'value1'", "true", "false"))
	bundle2 := b.AddBundle()
	service2 := b.AddService(bundle2, b.Criteria("label2 == 'value2'", "true", "false"))
	b.AddBundleComponent(bundle1, b.ServiceComponent(service2))

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claim with full labels (should be resolved successfully)
	c1 := b.AddClaim(b.AddUser(), service1)
	c1.Labels["label1"] = "value1"
	c1.Labels["label2"] = "value2"

	// add claim with partial labels (should not resolved)
	c2 := b.AddClaim(b.AddUser(), service1)
	c2.Labels["label1"] = "value1"

	// policy resolution should be completed successfully
	resolvePolicy(t, b, []verifyClaim{
		{claim: c1, resolved: true},
		{claim: c2, resolved: false, logMessage: "unable to find matching context"},
	})
}

func TestPolicyResolverCalculatedLabels(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// first service adds a label 'labelExtra1'
	bundle1 := b.AddBundle()
	service1 := b.AddService(bundle1, b.Criteria("label1 == 'value1'", "true", "false"))
	service1.ChangeLabels = lang.NewLabelOperationsSetSingleLabel("labelExtra1", "labelValue1")

	// second service adds a label 'labelExtra2' and removes 'label3'
	bundle2 := b.AddBundle()
	service2 := b.AddService(bundle2, b.Criteria("label2 == 'value2'", "true", "false"))
	service2.ChangeLabels = lang.NewLabelOperations(
		map[string]string{"labelExtra2": "labelValue2"},
		map[string]string{"label3": ""},
	)
	service2.Contexts[0].ChangeLabels = lang.NewLabelOperationsSetSingleLabel("labelExtra3", "labelValue3")
	b.AddBundleComponent(bundle1, b.ServiceComponent(service2))

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claim with labels 'label1', 'label2' and 'label3'
	claim := b.AddClaim(b.AddUser(), service1)
	claim.Labels["label1"] = "value1"
	claim.Labels["label2"] = "value2"
	claim.Labels["label3"] = "value3"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, []verifyClaim{
		{claim: claim, resolved: true},
	})

	// check labels for the end bundle (bundle2/service2)
	bundleInstance := getInstanceByParams(t, cluster, "k8ns", service2, service2.Contexts[0], nil, bundle2, nil, resolution)
	labels := bundleInstance.CalculatedLabels.Labels

	assert.Equal(t, "value1", labels["label1"], "Label 'label1=value1' should be carried from claim all the way through the policy")
	assert.Equal(t, "value2", labels["label2"], "Label 'label2=value2' should be carried from claim all the way through the policy")
	assert.NotContains(t, labels, "label3", "Label 'label3' should be removed")

	assert.Equal(t, "labelValue1", labels["labelExtra1"], "Label 'labelExtra1' should be added on service match")
	assert.Equal(t, "labelValue2", labels["labelExtra2"], "Label 'labelExtra2' should be added on service match")
	assert.Equal(t, "labelValue3", labels["labelExtra3"], "Label 'labelExtra3' should be added on context match")

	assert.Equal(t, cluster.Name, labels[lang.LabelTarget], "Label 'cluster' should be set")
}

func TestPolicyResolverCodeAndDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle with 2 components and multiple parameters
	bundle := b.AddBundle()
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
	b.AddBundleComponent(bundle, component1)
	b.AddBundleComponent(bundle, component2)

	service := b.AddService(bundle, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claims which feed conflicting labels into a given component
	claim := b.AddClaim(b.AddUser(), service)

	// policy should be resolved successfully
	resolution := resolvePolicy(t, b, []verifyClaim{
		{claim: claim, resolved: true},
	})

	// check discovery parameters of component 1
	instance1 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[0], nil, bundle, component1, resolution)
	assert.Regexp(t, "^component1-(.+)$", instance1.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// check discovery parameters of component 2
	instance2 := getInstanceByParams(t, cluster, "k8ns", service, service.Contexts[0], nil, bundle, component2, resolution)
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

func TestPolicyResolverClaimWithNonExistingUser(t *testing.T) {
	b := builder.NewPolicyBuilder()
	bundle := b.AddBundle()
	service := b.AddService(bundle, b.CriteriaTrue())
	user := &lang.User{Name: "non-existing-user-123456789"}
	claim := b.AddClaim(user, service)

	// claim declared by non-existing consumer should result in claim error
	resolvePolicy(t, b, []verifyClaim{
		{claim: claim, resolved: false, logMessage: "non-existing user"},
	})
}

func TestPolicyResolverConflictingCodeParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle which uses label in its code parameters
	bundle := b.AddBundle()
	b.AddBundleComponent(bundle,
		b.CodeComponent(
			util.NestedParameterMap{"address": "{{ .Labels.deplabel }}"},
			nil,
		),
	)
	service := b.AddService(bundle, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claims which feed conflicting labels into a given component
	c1 := b.AddClaim(b.AddUser(), service)
	c1.Labels["deplabel"] = "1"
	c2 := b.AddClaim(b.AddUser(), service)
	c2.Labels["deplabel"] = "2"

	// policy resolution with conflicting code parameters should result in an error
	resolvePolicy(t, b, []verifyClaim{
		{claim: c1, resolved: false, logMessage: "conflicting code parameters"},
		{claim: c2, resolved: false, logMessage: "conflicting code parameters"},
	})
}

func TestPolicyResolverConflictingDiscoveryParams(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle which uses label in its discovery parameters
	bundle := b.AddBundle()
	b.AddBundleComponent(bundle,
		b.CodeComponent(
			nil,
			util.NestedParameterMap{"address": "{{ .Labels.deplabel }}"},
		),
	)

	service := b.AddService(bundle, b.CriteriaTrue())
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claims which feed conflicting labels into a given component
	c1 := b.AddClaim(b.AddUser(), service)
	c1.Labels["deplabel"] = "1"
	c2 := b.AddClaim(b.AddUser(), service)
	c2.Labels["deplabel"] = "2"

	// policy resolution with conflicting discovery parameters should result in an error
	resolvePolicy(t, b, []verifyClaim{
		{claim: c1, resolved: false, logMessage: "conflicting discovery parameters"},
		{claim: c2, resolved: false, logMessage: "conflicting discovery parameters"},
	})
}

func TestPolicyResolverBundleLoop(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create 3 bundles
	bundle1 := b.AddBundle()
	service1 := b.AddService(bundle1, b.CriteriaTrue())
	bundle2 := b.AddBundle()
	service2 := b.AddService(bundle2, b.CriteriaTrue())
	bundle3 := b.AddBundle()
	service3 := b.AddService(bundle3, b.CriteriaTrue())

	// create bundle-level cycle
	b.AddBundleComponent(bundle1, b.ServiceComponent(service2))
	b.AddBundleComponent(bundle2, b.ServiceComponent(service3))
	b.AddBundleComponent(bundle3, b.ServiceComponent(service1))

	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add claim
	claim := b.AddClaim(b.AddUser(), service1)

	// policy resolution with bundle cycle should should result in an error
	resolvePolicy(t, b, []verifyClaim{
		{claim: claim, resolved: false, logMessage: "bundle cycle detected"},
	})
}

func TestPolicyResolverPickClusterViaRules(t *testing.T) {
	b := builder.NewPolicyBuilder()

	// create a bundle which can be deployed to different clusters
	bundle := b.AddBundle()
	b.AddBundleComponent(bundle,
		b.CodeComponent(
			util.NestedParameterMap{"debug": "{{ .Labels.target }}"},
			nil,
		),
	)
	service := b.AddService(bundle, b.CriteriaTrue())

	// add rules, which say to deploy to different clusters based on label value
	cluster1 := b.AddCluster()
	cluster2 := b.AddCluster()
	b.AddRule(b.Criteria("label1 == 'value1'", "true", "false"), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster1.Name)))
	b.AddRule(b.Criteria("label2 == 'value2'", "true", "false"), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster2.Name)))

	// add claims
	c1 := b.AddClaim(b.AddUser(), service)
	c1.Labels["label1"] = "value1"
	c2 := b.AddClaim(b.AddUser(), service)
	c2.Labels["label2"] = "value2"

	// policy resolution should be completed successfully
	resolution := resolvePolicy(t, b, []verifyClaim{
		{claim: c1, resolved: true},
		{claim: c2, resolved: true},
	})

	// check that both claims got resolved and got placed in different clusters
	instance1 := resolution.ComponentInstanceMap[resolution.GetClaimResolution(c1).ComponentInstanceKey]
	instance2 := resolution.ComponentInstanceMap[resolution.GetClaimResolution(c2).ComponentInstanceKey]
	assert.Equal(t, cluster1.Name, instance1.CalculatedLabels.Labels[lang.LabelTarget], "Cluster should be set correctly via rules")
	assert.Equal(t, cluster2.Name, instance2.CalculatedLabels.Labels[lang.LabelTarget], "Cluster should be set correctly via rules")
}

func TestPolicyResolverInternalPanic(t *testing.T) {
	b := builder.NewPolicyBuilder()
	b.PanicWhenLoadingUsers()

	// create a bundle with two contexts within a service
	bundle := b.AddBundle()
	b.AddBundleComponent(bundle, b.CodeComponent(nil, nil))
	service := b.AddService(bundle, b.CriteriaTrue())

	// add rule to set cluster
	cluster := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

	// add multiple claims
	expected := []verifyClaim{}
	for i := 0; i < 10; i++ {
		claim := b.AddClaim(b.AddUser(), service)
		expected = append(expected, verifyClaim{
			claim: claim, resolved: false, logMessage: "panic from mock user loader",
		})
	}

	// policy resolution should result in an error
	resolvePolicy(t, b, expected)
}

func TestPolicyResolverAllocationKeys(t *testing.T) {
	b := builder.NewPolicyBuilder()

	tc := map[string]bool{
		"fixedValue":       true,
		"{{ .User.Name }}": false,
		"{{ .Claim.ID }}":  false,
	}

	for allocationKey, sameInstance := range tc {
		// create a bundle and a service
		bundle := b.AddBundle()
		b.AddBundleComponent(bundle, b.CodeComponent(nil, nil))
		service := b.AddService(bundle, b.CriteriaTrue())
		service.Contexts[0].Allocation.Keys = b.AllocationKeys(allocationKey)

		// add rule to set cluster
		cluster := b.AddCluster()
		b.AddRule(b.Criteria(
			"Bundle.Namespace == 'main' && Claim.Namespace == 'main'",
			"true",
			"false",
		), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, cluster.Name)))

		// add claim (should be resolved to the first context)
		c1 := b.AddClaim(b.AddUser(), service)

		// add claim (should be resolved to the second context)
		c2 := b.AddClaim(b.AddUser(), service)

		// policy resolution should be completed successfully
		resolution := resolvePolicy(t, b, []verifyClaim{
			{claim: c1, resolved: true},
			{claim: c2, resolved: true},
		})

		// make sure claims point to different instances
		if sameInstance {
			assert.Equal(t, resolution.GetClaimResolution(c1).ComponentInstanceKey, resolution.GetClaimResolution(c2).ComponentInstanceKey, "Claims should point to same instance")
		} else {
			assert.NotEqual(t, resolution.GetClaimResolution(c1).ComponentInstanceKey, resolution.GetClaimResolution(c2).ComponentInstanceKey, "Claims should point to different instances")
		}
	}
}

/*
	Helpers
*/

type verifyClaim struct {
	claim      *lang.Claim
	resolved   bool
	logMessage string
}

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder, expected []verifyClaim) *PolicyResolution {
	t.Helper()
	eventLog := event.NewLog(logrus.DebugLevel, "test-resolve")
	resolver := NewPolicyResolver(builder.Policy(), builder.External(), eventLog)
	result := resolver.ResolveAllClaims()

	// check status of all claims
	for _, check := range expected {
		// check status first
		if !assert.Equal(t, check.resolved, result.GetClaimResolution(check.claim).Resolved, "Claim resolution status should be correct for %v", check.claim) {
			// print log into stdout and exit
			hook := event.NewHookConsole(logrus.DebugLevel)
			eventLog.Save(hook)
			t.FailNow()
			return nil
		}
		// check for log message
		if len(check.logMessage) > 0 {
			verifier := event.NewLogVerifier(check.logMessage, !check.resolved)
			resolver.eventLog.Save(verifier)
			if !assert.True(t, verifier.MatchedErrorsCount() > 0, "Event log should have an error message containing words: %s", check.logMessage) {
				hook := event.NewHookConsole(logrus.DebugLevel)
				resolver.eventLog.Save(hook)
				t.FailNow()
			}
		}
	}

	return result
}

func getInstanceByParams(t *testing.T, cluster *lang.Cluster, namespace string, service *lang.Service, context *lang.Context, allocationKeysResolved []string, bundle *lang.Bundle, component *lang.BundleComponent, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := NewComponentInstanceKey(cluster, namespace, service, context, allocationKeysResolved, bundle, component)
	instance, ok := resolution.ComponentInstanceMap[key.GetKey()]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key.GetKey()) {
		t.FailNow()
	}
	return instance
}
