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

// AddClaim creates a new claim and adds it to the policy
func (builder *PolicyBuilder) AddClaim(user *lang.User, service *lang.Service) *lang.Claim {
	result := &lang.Claim{
		TypeKind: lang.ClaimObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		User:    strings.ToUpper(user.Name), // we can refer to user using any case, since user name is not case sensitive
		Service: service.Namespace + "/" + service.Name,
		Labels:  make(map[string]string),
	}

	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddUser creates a new user who can consume bundles from the 'main' namespace and adds it to the policy
func (builder *PolicyBuilder) AddUser() *lang.User {
	result := &lang.User{
		Name:        util.RandomID(builder.random, idLength),
		Labels:      map[string]string{},
		DomainAdmin: true, // this will ensure that this user can consume bundles
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

// AddBundle creates a new bundle and adds it to the policy
func (builder *PolicyBuilder) AddBundle() *lang.Bundle {
	result := &lang.Bundle{
		TypeKind: lang.BundleObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
	}
	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddService creates a new service for a given bundle and adds it to the policy
func (builder *PolicyBuilder) AddService(bundle *lang.Bundle, criteria *lang.Criteria) *lang.Service {
	result := &lang.Service{
		TypeKind: lang.ServiceObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: builder.namespace,
			Name:      util.RandomID(builder.random, idLength),
		},
		Contexts: []*lang.Context{{
			Name:     util.RandomID(builder.random, idLength),
			Criteria: criteria,
			Allocation: &lang.Allocation{
				Bundle: bundle.Name,
			},
		}},
	}
	builder.addObject(builder.domainAdminView, result)
	return result
}

// AddServiceMultipleContexts creates service with multiple contexts for a given bundle and adds it to the policy
func (builder *PolicyBuilder) AddServiceMultipleContexts(bundle *lang.Bundle, criteriaArray ...*lang.Criteria) *lang.Service {
	result := &lang.Service{
		TypeKind: lang.ServiceObject.GetTypeKind(),
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
					Bundle: bundle.Name,
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

// UnknownComponent creates an unknown component for a bundle (not code and not service)
func (builder *PolicyBuilder) UnknownComponent() *lang.BundleComponent {
	return &lang.BundleComponent{
		Name: util.RandomID(builder.random, idLength),
	}
}

// CodeComponent creates a new code component for a bundle
func (builder *PolicyBuilder) CodeComponent(codeParams util.NestedParameterMap, discoveryParams util.NestedParameterMap) *lang.BundleComponent {
	return &lang.BundleComponent{
		Name: util.RandomID(builder.random, idLength),
		Code: &lang.Code{
			Type:   "helm",
			Params: codeParams,
		},
		Discovery: discoveryParams,
	}
}

// ServiceComponent creates a new service component for a bundle
func (builder *PolicyBuilder) ServiceComponent(service *lang.Service) *lang.BundleComponent {
	return &lang.BundleComponent{
		Name:    util.RandomID(builder.random, idLength),
		Service: service.Namespace + "/" + service.Name,
	}
}

// AddBundleComponent adds a given bundle component to the bundle
func (builder *PolicyBuilder) AddBundleComponent(bundle *lang.Bundle, component *lang.BundleComponent) *lang.BundleComponent {
	bundle.Components = append(bundle.Components, component)
	return component
}

// AddComponentClaim adds a component claim on another component
func (builder *PolicyBuilder) AddComponentClaim(component *lang.BundleComponent, dependsOn *lang.BundleComponent) {
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
