package builder

import (
	"math/rand"
	"strings"

	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/external/secrets"
	"github.com/Aptomi/aptomi/pkg/external/users"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

var randSeed = int64(239)
var idLength = 16

// PolicyBuilder provides simple and easy-to-use way to construct a complete Policy for Aptomi
// in the source code. It is primarily used in unit tests.
// When objects are created/added by the policy builder, they get created in the specified namespace
// with randomly generated IDs/names, so that a user doesn't have to specify them.
type PolicyBuilder struct {
	random    *rand.Rand
	namespace string
	policy    *lang.Policy
	users     *users.UserLoaderMock
	secrets   *secrets.SecretLoaderMock

	domainAdmin     *lang.User
	domainAdminView *lang.PolicyView
}

// NewPolicyBuilder creates a new PolicyBuilder with a default "main" namespace
func NewPolicyBuilder() *PolicyBuilder {
	return NewPolicyBuilderWithNS("main")
}

// NewPolicyBuilderWithNS creates a new PolicyBuilder
func NewPolicyBuilderWithNS(namespace string) *PolicyBuilder {
	result := &PolicyBuilder{
		random:    rand.New(rand.NewSource(randSeed)),
		namespace: namespace,
		policy:    lang.NewPolicy(),
		users:     users.NewUserLoaderMock(),
		secrets:   secrets.NewSecretLoaderMock(),
	}

	result.domainAdmin = result.AddUserDomainAdmin()
	result.domainAdminView = result.policy.View(result.domainAdmin)

	return result
}

// SwitchNamespace switches the current namespace where objects will be generated
func (builder *PolicyBuilder) SwitchNamespace(namespace string) {
	builder.namespace = namespace
}

// AddDependency creates a new dependency and adds it to the policy
func (builder *PolicyBuilder) AddDependency(user *lang.User, contract *lang.Contract) *lang.Dependency {
	result := &lang.Dependency{
		TypeKind: lang.DependencyObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		User:     strings.ToUpper(user.Name), // we can refer to user using any case, since user name is not case sensitive
		Contract: contract.Namespace + "/" + contract.Name,
		Labels:   make(map[string]string),
	}

	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddUser creates a new user who can consume services from the 'main' namespace and adds it to the policy
func (builder *PolicyBuilder) AddUser() *lang.User {
	result := &lang.User{
		Name:        util.RandomID(builder.random, idLength),
		Labels:      map[string]string{},
		DomainAdmin: true, // this will ensure that this user can consume services
	}
	builder.users.AddUser(result)
	return result
}

// PanicWhenLoadingUsers tells mock user loader to start panicking when loading users
func (builder *PolicyBuilder) PanicWhenLoadingUsers() {
	builder.users.SetPanic(true)
}

// AddUserDomainAdmin creates a new user who is a domain admin and adds it to the policy
func (builder *PolicyBuilder) AddUserDomainAdmin() *lang.User {
	return builder.AddUser()
}

// AddService creates a new service and adds it to the policy
func (builder *PolicyBuilder) AddService() *lang.Service {
	result := &lang.Service{
		TypeKind: lang.ServiceObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
	}
	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddContract creates a new contract for a given service and adds it to the policy
func (builder *PolicyBuilder) AddContract(service *lang.Service, criteria *lang.Criteria) *lang.Contract {
	result := &lang.Contract{
		TypeKind: lang.ContractObject.GetTypeKind(),
		Metadata: lang.Metadata{
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
	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddContractMultipleContexts creates contract with multiple contexts for a given service and adds it to the policy
func (builder *PolicyBuilder) AddContractMultipleContexts(service *lang.Service, criteriaArray ...*lang.Criteria) *lang.Contract {
	result := &lang.Contract{
		TypeKind: lang.ContractObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
	}
	for _, criteria := range criteriaArray {
		result.Contexts = append(result.Contexts,
			&lang.Context{
				Name:     util.RandomID(builder.random, idLength),
				Criteria: criteria,
				Allocation: &lang.Allocation{
					Service: service.Name,
				},
			},
		)
	}

	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddRule creates a new rule and adds it to the policy
func (builder *PolicyBuilder) AddRule(criteria *lang.Criteria, actions *lang.RuleActions) *lang.Rule {
	result := &lang.Rule{
		TypeKind: lang.RuleObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		Weight:   len(builder.policy.GetObjectsByKind(lang.RuleObject.Kind)),
		Criteria: criteria,
		Actions:  actions,
	}
	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddCluster creates a new cluster and adds it to the policy
func (builder *PolicyBuilder) AddCluster() *lang.Cluster {
	result := &lang.Cluster{
		TypeKind: lang.ClusterObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: runtime.SystemNS,
			Name:      util.RandomID(builder.random, idLength),
		},
		Type: "kubernetes",
		Config: struct {
			Namespace        string
			DefaultNamespace string
		}{
			Namespace:        "default",
			DefaultNamespace: "k8ns",
		},
	}
	builder.addObject(builder.domainAdminView, result)
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
func (builder *PolicyBuilder) AllocationKeys(key ...string) []string {
	return key
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
			Type:   "helm",
			Params: codeParams,
		},
		Discovery: discoveryParams,
	}
}

// ContractComponent creates a new contract component for a service
func (builder *PolicyBuilder) ContractComponent(contract *lang.Contract) *lang.ServiceComponent {
	return &lang.ServiceComponent{
		Name:     util.RandomID(builder.random, idLength),
		Contract: contract.Namespace + "/" + contract.Name,
	}
}

// AddServiceComponent adds a given service component to the service
func (builder *PolicyBuilder) AddServiceComponent(service *lang.Service, component *lang.ServiceComponent) *lang.ServiceComponent {
	service.Components = append(service.Components, component)
	return component
}

// AddComponentDependency adds a component dependency on another component
func (builder *PolicyBuilder) AddComponentDependency(component *lang.ServiceComponent, dependsOn *lang.ServiceComponent) {
	component.Dependencies = append(component.Dependencies, dependsOn.Name)
}

// RuleActions creates a new RuleActions object
func (builder *PolicyBuilder) RuleActions(labelOps lang.LabelOperations) *lang.RuleActions {
	result := &lang.RuleActions{}
	if labelOps != nil {
		result.ChangeLabels = labelOps
	}
	return result
}

// Policy returns the generated policy
func (builder *PolicyBuilder) Policy() *lang.Policy {
	err := builder.policy.Validate()
	if err != nil {
		panic(err)
	}
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

// Internal function to add objects to the policy
func (builder *PolicyBuilder) addObject(view *lang.PolicyView, obj lang.Base) {
	err := view.AddObject(obj)
	if err != nil {
		panic(err)
	}
}
