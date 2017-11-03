package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
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
	return true
}

func TestPolicyValidationService(t *testing.T) {
	// Service (Identifiers & Labels)
	runValidationTests(t, ResSuccess, true, []object.Base{
		makeService("test", 0),
		makeService("good_name", Empty),
		makeService("excellent-name_239", Nil),
	})
	runValidationTests(t, ResFailure, true, []object.Base{
		makeService("_invalid", 0),
		makeService("12-invalid", 0),
		makeService("invalid-n#ame_239", 0),
		makeService("invalid-n$ame_239", 0),
		makeService("valid", Invalid),
	})

	// Service Components
	contract := makeContract("contract", 0, "")
	componentTestsPass := [][]*ServiceComponent{
		makeServiceComponents(1, contract.Name, Nil, 0),
		makeServiceComponents(2, contract.Name, Nil, 0),
		makeServiceComponents(3, "", 0, 1),
		makeServiceComponents(4, "", 1, 1),
	}
	for _, components := range componentTestsPass {
		service := makeService("service", Empty)
		service.Components = components
		runValidationTests(t, ResSuccess, false, []object.Base{service, contract})
	}
	componentTestsFail := [][]*ServiceComponent{
		makeServiceComponents(1, contract.Name+"extra", Nil, 0),
		makeServiceComponents(1, contract.Name, Empty, 0),
		makeServiceComponents(1, "", Empty, 0),
		makeServiceComponents(1, "", Nil, 0),
		makeServiceComponents(1, "", Invalid, 0),
		makeServiceComponents(1, "", Invalid-1, 0),
		makeServiceComponents(1, contract.Name, Nil, Invalid),
		duplicateNames(makeServiceComponents(10, "", 1, 1)),
		dependenciesInvalid(makeServiceComponents(10, "", 1, 1)),
		dependenciesCycle(makeServiceComponents(10, "", 1, 1)),
	}
	for _, components := range componentTestsFail {
		service := makeService("service", Empty)
		service.Components = components
		runValidationTests(t, ResFailure, false, []object.Base{service, contract})
	}
}

func TestPolicyValidationContract(t *testing.T) {
	// Contract (Identifiers & Label Operations)
	runValidationTests(t, ResSuccess, true, []object.Base{
		makeContract("test", 0, ""),
		makeContract("test", 1, ""),
		makeContract("test", Empty, ""),
		makeContract("test", Nil, ""),
	})
	runValidationTests(t, ResFailure, true, []object.Base{
		makeContract("_invalid", 0, ""),
		makeContract("valid", Invalid, ""),
	})

	// Contract should point to an existing service
	runValidationTests(t, ResSuccess, false, []object.Base{
		makeService("service", Empty),
		makeContract("test1", 0, "service"),
		makeContract("test2", 1, "service"),
	})
	runValidationTests(t, ResFailure, false, []object.Base{
		makeService("service", Empty),
		makeContract("test1", 0, "service-unknown"),
	})
}

func TestPolicyValidationDependency(t *testing.T) {
	// Dependency should point to an existing contract
	runValidationTests(t, ResSuccess, false, []object.Base{
		makeContract("contract", 0, ""),
		makeDependency("contract"),
	})
	runValidationTests(t, ResFailure, false, []object.Base{
		makeContract("contract", 0, ""),
		makeDependency("contract-unknown"),
	})
}

func TestPolicyValidationRule(t *testing.T) {
	// Rules (Expressions & Actions)
	runValidationTests(t, ResSuccess, true, []object.Base{
		makeRule(1, "true", 0, "labelName"),
		makeRule(20, "", 1, Reject),
		makeRule(100, "specialname + specialvalue == 'b'", 2, Reject),
	})
	runValidationTests(t, ResFailure, true, []object.Base{
		makeRule(-1, "true", 0, "labelName"),                               // negative weight
		makeRule(100, "specialname + '123')(((", 0, "labelName"),           // bad expression
		makeRule(100, "true", Empty, ""),                                   // no actions specified
		makeRule(100, "true", Nil, ""),                                     // actions = nil
		makeRule(100, "specialname + specialvalue == 'b'", 2, "notreject"), // action is not (allow, reject)
	})
}

func TestPolicyValidationACLRule(t *testing.T) {
	// Rules (Expressions & Actions)
	runValidationTests(t, ResSuccess, true, []object.Base{
		makeACLRule(0),
	})
	runValidationTests(t, ResFailure, true, []object.Base{
		makeACLRule(Empty),
		makeACLRule(Nil),
		makeACLRule(Invalid),
	})
}

func TestPolicyValidationCluster(t *testing.T) {
	// Clusters (Identifiers & Config)
	runValidationTests(t, ResSuccess, true, []object.Base{
		makeCluster("kubernetes"),
	})
	runValidationTests(t, ResFailure, true, []object.Base{
		makeCluster("unknown"),
	})
}

func runValidationTests(t *testing.T, result int, every bool, objects []object.Base) {
	t.Helper()

	if every {
		// one by one
		for _, obj := range objects {
			policy := NewPolicy()
			err := policy.AddObject(obj)
			assert.NoError(t, err, "Unable to add object to policy: %s", obj)
			validatePolicy(t, result, []object.Base{obj}, policy)
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

func validatePolicy(t *testing.T, result int, objects []object.Base, policy *Policy) {
	t.Helper()
	errValidate := policy.Validate()

	var failed bool
	if result == ResSuccess {
		failed = !assert.NoError(t, errValidate, "Policy validation should succeed. Objects: \n%s", yaml.SerializeObject(objects))
	} else {
		failed = !assert.Error(t, errValidate, "Policy validation should fail. Objects: \n%s", yaml.SerializeObject(objects)) // nolint: vet
	}

	if errValidate != nil && (displayErrorMessages() || failed) {
		fmt.Println(errValidate)
	}
}

func makeRule(weight int, expr string, actionNum int, actionKey string) *Rule {
	rule := &Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
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
		rule.Actions = &RuleActions{Dependency: DependencyAction(actionKey)}
	case 2:
		rule.Actions = &RuleActions{Ingress: IngressAction(actionKey)}
	case Empty:
		rule.Actions = &RuleActions{}
	case Nil:
		// no actions defined, nil
	}

	return rule
}

func makeACLRule(actionNum int) *Rule {
	rule := &Rule{
		Metadata: Metadata{
			Kind:      ACLRuleObject.Kind,
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
		rule.Actions = &RuleActions{AddRole: map[string]string{domainAdmin.ID: namespaceAll, serviceConsumer.ID: "main1, main2 ,main3,main4"}}
	case Empty:
		rule.Actions = &RuleActions{}
	case Nil:
		// no actions defined, nil
	case Invalid:
		// invalid action tied to an unknown role
		rule.Actions = &RuleActions{AddRole: map[string]string{"unknown": "main1"}}
	}

	return rule
}

func makeContract(name string, labelOpsNum int, pointToService string) *Contract {
	contract := &Contract{
		Metadata: Metadata{
			Kind:      ContractObject.Kind,
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

	if len(pointToService) > 0 {
		contract.Contexts = []*Context{
			{
				Name: "context",
				Allocation: &Allocation{
					Service: pointToService,
					Keys:    []string{"simple", "{{ .Labels.cluster }}"},
				},
			},
		}
	}

	return contract
}

func makeCluster(clusterType string) *Cluster {
	return &Cluster{
		Metadata: Metadata{
			Kind:      ClusterObject.Kind,
			Namespace: object.SystemNS,
			Name:      "cluster",
		},
		Type: clusterType,
		Config: ClusterConfig{
			KubeContext:     "value",
			TillerNamespace: "value",
			Namespace:       "value",
		},
	}
}

func makeService(name string, labelNum int) *Service {
	service := &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      name,
		},
	}
	switch labelNum {
	case 0:
		service.Labels = map[string]string{"name": "value"}
	case Empty:
		// no labels defined, empty
		service.Labels = make(map[string]string)
	case Nil:
		// no labels defined, nil
	case Invalid:
		// invalid labels
		service.Labels = map[string]string{"$@#$%^&": "value"}
	}
	return service
}

func makeDependency(contract string) *Dependency {
	dependency := &Dependency{
		Metadata: Metadata{
			Kind:      DependencyObject.Kind,
			Namespace: "main",
			Name:      "dependency",
		},
		User:     "user",
		Contract: contract,
	}
	return dependency
}

func makeServiceComponents(count int, contract string, codeNum int, discoveryNum int) []*ServiceComponent {
	result := make([]*ServiceComponent, count)
	for i := 0; i < count; i++ {
		component := &ServiceComponent{
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
				Params: util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "{{ .Labels.cluster }}"}},
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
				Params: util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "{{ .Labels.cluster }}"}},
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
			component.Discovery = util.NestedParameterMap{"a": "aValue", "nested": util.NestedParameterMap{"c": "{{ .Labels.cluster }}"}}
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

func duplicateNames(components []*ServiceComponent) []*ServiceComponent {
	for _, component := range components {
		component.Name = "name"
	}
	return components
}

func dependenciesInvalid(components []*ServiceComponent) []*ServiceComponent {
	for _, component := range components {
		component.Dependencies = []string{"invalid"}
	}
	return components
}

func dependenciesCycle(components []*ServiceComponent) []*ServiceComponent {
	for _, component := range components {
		component.Dependencies = []string{component.Name}
	}
	return components
}
