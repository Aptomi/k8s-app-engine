package builder

import (
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"math/rand"
)

var randSeed = int64(239)
var idLength = 16

// PolicyBuilder is a utility struct to help build policy objects
// It is primarily used in unit tests
type PolicyBuilder struct {
	random    *rand.Rand
	namespace string
	policy    *lang.Policy
	users     *users.UserLoaderMock
	secrets   *secrets.SecretLoaderMock
}

// NewPolicyBuilder creates a new PolicyBuilder with a default "main" namespace
func NewPolicyBuilder() *PolicyBuilder {
	return NewPolicyBuilderWithNS("main")
}

// NewPolicyBuilderWithNS creates a new PolicyBuilder
func NewPolicyBuilderWithNS(namespace string) *PolicyBuilder {
	return &PolicyBuilder{
		random:    rand.New(rand.NewSource(randSeed)),
		namespace: namespace,
		policy:    lang.NewPolicy(),
		users:     users.NewUserLoaderMock(),
		secrets:   secrets.NewSecretLoaderMock(),
	}
}

// AddDependency creates a new dependency and adds it to the policy
func (builder *PolicyBuilder) AddDependency(user *lang.User, contract *lang.Contract) *lang.Dependency {
	result := &lang.Dependency{
		Metadata: lang.Metadata{
			Kind:      lang.DependencyObject.Kind,
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		UserID:   user.ID,
		Contract: contract.Name,
		Labels:   make(map[string]string),
	}

	builder.policy.AddObject(result)
	return result
}

// AddUser creates a new user and adds it to the policy
func (builder *PolicyBuilder) AddUser() *lang.User {
	result := &lang.User{
		ID:     util.RandomID(builder.random, idLength),
		Name:   util.RandomID(builder.random, idLength),
		Labels: make(map[string]string),
	}
	builder.users.AddUser(result)
	return result
}

// AddService creates a new service and adds it to the policy
func (builder *PolicyBuilder) AddService(owner *lang.User) *lang.Service {
	var ownerID = ""
	if owner != nil {
		ownerID = owner.ID
	}
	result := &lang.Service{
		Metadata: lang.Metadata{
			Kind:      lang.ServiceObject.Kind,
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		Owner: ownerID,
	}
	builder.policy.AddObject(result)
	return result
}

// AddContract creates a new contract for a given service and adds it to the policy
func (builder *PolicyBuilder) AddContract(service *lang.Service, criteria *lang.Criteria) *lang.Contract {
	result := &lang.Contract{
		Metadata: lang.Metadata{
			Kind:      lang.ContractObject.Kind,
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		Contexts: []*lang.Context{{
			Name:     util.RandomID(builder.random, idLength),
			Criteria: criteria,
			Allocation: &lang.Allocation{
				Service: service.Name,
			},
		}},
	}
	builder.policy.AddObject(result)
	return result
}

// AddRule creates a new rule and adds it to the policy
func (builder *PolicyBuilder) AddRule(criteria *lang.Criteria, actions *lang.RuleActions) *lang.Rule {
	result := &lang.Rule{
		Metadata: lang.Metadata{
			Kind:      lang.RuleObject.Kind,
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		Weight:   len(builder.policy.GetObjectsByKind(lang.RuleObject.Kind)),
		Criteria: criteria,
		Actions:  actions,
	}
	builder.policy.AddObject(result)
	return result
}

// AddCluster creates a new cluster and adds it to the policy
func (builder *PolicyBuilder) AddCluster() *lang.Cluster {
	result := &lang.Cluster{
		Metadata: lang.Metadata{
			Kind:      lang.ClusterObject.Kind,
			Namespace: object.SystemNS,
			Name:      util.RandomID(builder.random, idLength),
		},
	}
	builder.policy.AddObject(result)
	return result
}

// Criteria creates a criteria with one require-all, one require-any, and one require-none
func (builder *PolicyBuilder) Criteria(all string, any string, none string) *lang.Criteria {
	return &lang.Criteria{
		RequireAll:  []string{all},
		RequireAny:  []string{any},
		RequireNone: []string{none},
	}
}

// CriteriaTrue creates a criteria which always evaluates to true
func (builder *PolicyBuilder) CriteriaTrue() *lang.Criteria {
	return &lang.Criteria{
		RequireAny: []string{"true"},
	}
}

// AllocationKeys creates allocation keys
func (builder *PolicyBuilder) AllocationKeys(key string) []string {
	return []string{key}
}

// UnknownComponent creates an unknown component for a service (not code and not contract)
func (builder *PolicyBuilder) UnknownComponent() *lang.ServiceComponent {
	return &lang.ServiceComponent{
		Name: util.RandomID(builder.random, idLength),
	}
}

// CodeComponent creates a new code component for a service
func (builder *PolicyBuilder) CodeComponent(codeParams util.NestedParameterMap, discoveryParams util.NestedParameterMap) *lang.ServiceComponent {
	return &lang.ServiceComponent{
		Name: util.RandomID(builder.random, idLength),
		Code: &lang.Code{
			Type:   "aptomi/code/unittests",
			Params: codeParams,
		},
		Discovery: discoveryParams,
	}
}

// ContractComponent creates a new contract component for a service
func (builder *PolicyBuilder) ContractComponent(contract *lang.Contract) *lang.ServiceComponent {
	return &lang.ServiceComponent{
		Name:     util.RandomID(builder.random, idLength),
		Contract: contract.Name,
	}
}

// AddServiceComponent adds a given service component to the service
func (builder *PolicyBuilder) AddServiceComponent(service *lang.Service, component *lang.ServiceComponent) {
	service.Components = append(service.Components, component)
}

// AddComponentDependency adds a component dependency on another component
func (builder *PolicyBuilder) AddComponentDependency(component *lang.ServiceComponent, dependsOn *lang.ServiceComponent) {
	component.Dependencies = append(component.Dependencies, dependsOn.Name)
}

// RuleActions creates a new RuleActions object
func (builder *PolicyBuilder) RuleActions(dependencyAction string, ingresAction string, labelOps lang.LabelOperations) *lang.RuleActions {
	result := &lang.RuleActions{}
	if len(dependencyAction) > 0 {
		result.Dependency = lang.DependencyAction(dependencyAction)
	}
	if len(ingresAction) > 0 {
		result.Ingress = lang.IngressAction(dependencyAction)
	}
	if labelOps != nil {
		result.ChangeLabels = lang.ChangeLabelsAction(labelOps)
	}
	return result
}

// Policy returns the generated policy
func (builder *PolicyBuilder) Policy() *lang.Policy {
	return builder.policy
}

// External returns the generated external data
func (builder *PolicyBuilder) External() *external.Data {
	return external.NewData(
		builder.users,
		builder.secrets,
	)
}

// Namespace returns the current namespace
func (builder *PolicyBuilder) Namespace() string {
	return builder.namespace
}
