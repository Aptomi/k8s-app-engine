package lang

import (
	"strconv"
	"testing"

	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
)

const (
	ResSuccess = iota
	ResFailure = iota
)

const (
	Empty   = -1
	Nil     = -2
	Invalid = -3
)

func displayErrorMessages() bool {
	return false
}

func TestPolicyValidationBundle(t *testing.T) {
	// Bundle (Identifiers & Labels)
	runValidationTests(t, ResSuccess, true, []Base{
		makeBundle("test", 0),
		makeBundle("good_name", Empty),
		makeBundle("excellent-name_239", Nil),
	})
	runValidationTests(t, ResFailure, true, []Base{
		makeBundle("_invalid", 0),
		makeBundle("12-invalid", 0),
		makeBundle("invalid-n#ame_239", 0),
		makeBundle("invalid-n$ame_239", 0),
		makeBundle("valid", Invalid),
	})

	// Bundle Components
	contract := makeContract("contract", 0, "")
	componentTestsPass := [][]*BundleComponent{
		makeBundleComponents(1, contract.Name, Nil, 0),
		makeBundleComponents(2, contract.Name, Nil, 0),
		makeBundleComponents(3, "", 0, 1),
		makeBundleComponents(4, "", 1, 1),
	}
	for _, components := range componentTestsPass {
		bundle := makeBundle("bundle", Empty)
		bundle.Components = components
		runValidationTests(t, ResSuccess, false, []Base{bundle, contract})
	}
	componentTestsFail := [][]*BundleComponent{
		makeBundleComponents(1, contract.Name+"extra", Nil, 0),
		makeBundleComponents(1, contract.Name, Empty, 0),
		makeBundleComponents(1, "", Empty, 0),
		makeBundleComponents(1, "", Nil, 0),
		makeBundleComponents(1, "", Invalid, 0),
		makeBundleComponents(1, "", Invalid-1, 0),
		makeBundleComponents(1, contract.Name, Nil, Invalid),
		duplicateNames(makeBundleComponents(10, "", 1, 1)),
		claimsInvalid(makeBundleComponents(10, "", 1, 1)),
		claimsCycle(makeBundleComponents(10, "", 1, 1)),
	}
	for _, components := range componentTestsFail {
		bundle := makeBundle("bundle", Empty)
		bundle.Components = components
		runValidationTests(t, ResFailure, false, []Base{bundle, contract})
	}
}

func TestPolicyValidationContract(t *testing.T) {
	// Contract (Identifiers & Label Operations & Allocation Keys)
	runValidationTests(t, ResSuccess, true, []Base{
		makeContract("test", 0, ""),
		makeContract("test", 1, ""),
		makeContract("test", Empty, ""),
		makeContract("test", Nil, ""),
	})
	runValidationTests(t, ResFailure, true, []Base{
		makeContract("_invalid", 0, ""),
		makeContract("valid", Invalid, ""),
	})

	// Contract should point to an existing bundle
	runValidationTests(t, ResSuccess, false, []Base{
		makeBundle("bundle", Empty),
		makeContract("test1", 0, "bundle"),
		makeContract("test2", 1, "bundle"),
	})
	runValidationTests(t, ResFailure, false, []Base{
		makeBundle("bundle", Empty),
		makeContract("test1", 0, "bundle-unknown"),
	})

	// Check allocation keys
	runValidationTests(t, ResFailure, false, []Base{
		makeBundle("bundle", Empty),
		invalidAllocationKeys(makeContract("test1", 0, "bundle")),
	})
}

func TestPolicyValidationClaim(t *testing.T) {
	// Claim should point to an existing contract
	runValidationTests(t, ResSuccess, false, []Base{
		makeContract("contract", 0, ""),
		makeClaim("contract"),
	})
	runValidationTests(t, ResFailure, false, []Base{
		makeContract("contract", 0, ""),
		makeClaim("contract-unknown"),
	})
}

func TestPolicyValidationRule(t *testing.T) {
	// Rules (Expressions & Actions)
	runValidationTests(t, ResSuccess, true, []Base{
		makeRule(1, "true", 0, "labelName"),
		makeRule(20, "", 1, Reject),
		makeRule(100, "specialname + specialvalue == 'b'", 2, Reject),
	})
	runValidationTests(t, ResFailure, true, []Base{
		makeRule(-1, "true", 0, "labelName"),                               // negative weight
		makeRule(100, "specialname + '123')(((", 0, "labelName"),           // bad expression
		makeRule(100, "true", Empty, ""),                                   // no actions specified
		makeRule(100, "true", Nil, ""),                                     // actions = nil
		makeRule(100, "specialname + specialvalue == 'b'", 2, "notreject"), // action is not (allow, reject)
	})
}

func TestPolicyValidationACLRule(t *testing.T) {
	// Rules (Expressions & Actions)
	runValidationTests(t, ResSuccess, true, []Base{
		makeACLRule(0),
	})
	runValidationTests(t, ResFailure, true, []Base{
		makeACLRule(Empty),
		makeACLRule(Nil),
		makeACLRule(Invalid),
	})
}

func TestPolicyValidationCluster(t *testing.T) {
	// Clusters (Identifiers & Config)
	runValidationTests(t, ResSuccess, true, []Base{
		makeCluster("kubernetes", runtime.SystemNS),
	})
	runValidationTests(t, ResFailure, true, []Base{
		makeCluster("unknown", runtime.SystemNS),
		makeCluster("kubernetes", "main"),
	})
}

func runValidationTests(t *testing.T, result int, every bool, objects []Base) {
	t.Helper()

	if every {
		// one by one
		for _, obj := range objects {
			policy := NewPolicy()
			err := policy.AddObject(obj)
			assert.NoError(t, err, "Unable to add object to policy: %s", obj)
			validatePolicy(t, result, []Base{obj}, policy)
		}
	} else {
		// all at once
		policy := NewPolicy()
		for _, obj := range objects {
			err := policy.AddObject(obj)
			assert.NoError(t, err, "Unable to add object to policy: %s", obj)
		}
		validatePolicy(t, result, objects, policy)
	}
}

func validatePolicy(t *testing.T, result int, objects []Base, policy *Policy) {
	t.Helper()
	errValidate := policy.Validate()

	var failed bool
	if result == ResSuccess {
		failed = !assert.NoError(t, errValidate, "Policy validation should succeed. Objects: \n%s", yaml.SerializeObject(objects))
	} else {
		failed = !assert.Error(t, errValidate, "Policy validation should fail. Objects: \n%s", yaml.SerializeObject(objects)) // nolint: vet
	}

	if errValidate != nil {
		if !assert.NotContains(t, errValidate.Error(), "Error:Field validation", "Policy validation error message is not human readable. Must define a translator") {
			t.Log(errValidate)
		}
	}

	if errValidate != nil && (displayErrorMessages() || failed) {
		t.Log(errValidate)
	}
}

func makeRule(weight int, expr string, actionNum int, actionKey string) *Rule {
	rule := &Rule{
		TypeKind: RuleObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "rule",
		},
		Weight: weight,
	}
	if len(expr) > 0 {
		rule.Criteria = &Criteria{
			RequireAll:  []string{"true"},
			RequireAny:  []string{"true", "true"},
			RequireNone: []string{expr},
		}
	}
	switch actionNum {
	case 0:
		rule.Actions = &RuleActions{ChangeLabels: NewLabelOperationsSetSingleLabel(actionKey, "value")}
	case 1:
		rule.Actions = &RuleActions{Claim: ClaimAction(actionKey)}
	case 2:
		rule.Actions = &RuleActions{Ingress: IngressAction(actionKey)}
	case Empty:
		rule.Actions = &RuleActions{}
	case Nil:
		// no actions defined, nil
	}

	return rule
}

func makeACLRule(actionNum int) *ACLRule {
	rule := &ACLRule{
		TypeKind: ACLRuleObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "rule",
		},
		Weight: 10,
	}
	rule.Criteria = &Criteria{
		RequireAll:  []string{"true"},
		RequireAny:  []string{"true", "true"},
		RequireNone: []string{"false", "false", "false"},
	}
	switch actionNum {
	case 0:
		rule.Actions = &ACLRuleActions{AddRole: map[string]string{DomainAdmin.ID: namespaceAll, ServiceConsumer.ID: "main1, main2 ,main3,main4"}}
	case Empty:
		rule.Actions = &ACLRuleActions{}
	case Nil:
		// no actions defined, nil
	case Invalid:
		// invalid action tied to an unknown role
		rule.Actions = &ACLRuleActions{AddRole: map[string]string{"unknown": "main1"}}
	}

	return rule
}

func makeContract(name string, labelOpsNum int, pointToBundle string) *Contract {
	contract := &Contract{
		TypeKind: ContractObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      name,
		},
	}
	switch labelOpsNum {
	case 0:
		contract.ChangeLabels = NewLabelOperationsSetSingleLabel("name", "value")
	case 1:
		contract.ChangeLabels = NewLabelOperations(map[string]string{"a": "a"}, map[string]string{"b": ""})
	case Empty:
		contract.ChangeLabels = LabelOperations{}
	case Nil:
		// no labels defined, nil
	case Invalid:
		contract.ChangeLabels = LabelOperations{"invalidOperation": map[string]string{"a": "a"}}
	}

	if len(pointToBundle) > 0 {
		contract.Contexts = []*Context{
			{
				Name: "context",
				Allocation: &Allocation{
					Bundle: pointToBundle,
					Keys:   []string{"simple", "{{ .Claim.ID }}"},
				},
			},
		}
	}

	return contract
}

func invalidAllocationKeys(contract *Contract) *Contract {
	for _, context := range contract.Contexts {
		context.Allocation.Keys = []string{"{{{ invalid"}
	}
	return contract
}

func makeCluster(clusterType, ns string) *Cluster {
	return &Cluster{
		TypeKind: ClusterObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: ns,
			Name:      "cluster",
		},
		Type:   clusterType,
		Config: "something",
	}
}

func makeBundle(name string, labelNum int) *Bundle {
	bundle := &Bundle{
		TypeKind: BundleObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      name,
		},
	}
	switch labelNum {
	case 0:
		bundle.Labels = map[string]string{"name": "value"}
	case Empty:
		// no labels defined, empty
		bundle.Labels = make(map[string]string)
	case Nil:
		// no labels defined, nil
	case Invalid:
		// invalid labels
		bundle.Labels = map[string]string{"$@#$%^&": "value"}
	}
	return bundle
}

func makeClaim(contract string) *Claim {
	claim := &Claim{
		TypeKind: ClaimObject.GetTypeKind(),
		Metadata: Metadata{
			Namespace: "main",
			Name:      "claim",
		},
		User:     "user",
		Contract: contract,
	}
	return claim
}

func makeBundleComponents(count int, contract string, codeNum int, discoveryNum int) []*BundleComponent {
	result := make([]*BundleComponent, count)
	for i := 0; i < count; i++ {
		component := &BundleComponent{
			Name: "component-" + strconv.Itoa(i),
		}
		if len(contract) > 0 {
			component.Contract = contract
		}
		switch codeNum {
		case 0:
			component.Code = &Code{
				Type:   "helm",
				Params: util.NestedParameterMap{},
			}
		case 1:
			component.Code = &Code{
				Type:   "helm",
				Params: util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "d"}},
			}
		case Empty:
			// no code defined, empty
			component.Code = &Code{}
		case Nil:
			// no code defined, nil
		case Invalid:
			// invalid code
			component.Code = &Code{
				Type:   "unknown",
				Params: util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "d"}},
			}
		case Invalid - 1:
			// invalid code
			component.Code = &Code{
				Type:   "helm",
				Params: util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "{{ broken___$$%@ }}"}},
			}
		}

		switch discoveryNum {
		case 0:
			component.Discovery = util.NestedParameterMap{"a": "b"}
		case 1:
			component.Discovery = util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "d"}}
		case Empty:
			// no discovery defined, empty
			component.Discovery = util.NestedParameterMap{}
		case Nil:
			// no discovery defined, nil
		case Invalid:
			// invalid discovery
			component.Discovery = util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "{{ broken___$$%@ }}"}}
		}

		for j := 0; j < i; j++ {
			component.Dependencies = append(component.Dependencies, "component-"+strconv.Itoa(j))
		}
		result[i] = component
	}
	return result
}

func duplicateNames(components []*BundleComponent) []*BundleComponent {
	for _, component := range components {
		component.Name = "name"
	}
	return components
}

func claimsInvalid(components []*BundleComponent) []*BundleComponent {
	for _, component := range components {
		component.Dependencies = []string{"invalid"}
	}
	return components
}

func claimsCycle(components []*BundleComponent) []*BundleComponent {
	for _, component := range components {
		component.Dependencies = []string{component.Name}
	}
	return components
}
