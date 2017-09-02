package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkEngine(b *testing.B) {
	t := &testing.T{}
	gen := NewPolicyGenerator(
		239,
		30,
		50, // 1000
		6,
		6,
		4,
		2,
		100,
		500,
	)
	desiredPolicy, externalData := gen.makePolicyAndExternalData()
	for i := 0; i < b.N; i++ {
		RunEngine(t, desiredPolicy, externalData)
	}
}

type PolicyGenerator struct {
	random                    *rand.Rand
	labels                    int
	services                  int
	serviceCodeComponents     int
	codeParams                int
	serviceDependencyMaxChain int
	contextsPerService        int
	users                     int
	dependencies              int

	generatedLabels    map[string]string
	generatedLabelKeys []string

	generatedServices []*language.Service
	policy            *language.PolicyNamespace
	externalData      *external.Data
}

func NewPolicyGenerator(randSeed int64, labels, services, serviceCodeComponents, codeParams, serviceDependencyMaxChain, contextsPerService, users int, dependencies int) *PolicyGenerator {
	return &PolicyGenerator{
		random:                    rand.New(rand.NewSource(randSeed)),
		labels:                    labels,
		services:                  services,
		serviceCodeComponents:     serviceCodeComponents,
		codeParams:                codeParams,
		serviceDependencyMaxChain: serviceDependencyMaxChain,
		contextsPerService:        contextsPerService,
		users:                     users,
		dependencies:              dependencies,
		policy:                    language.NewPolicyNamespace(),
	}
}

func (gen *PolicyGenerator) makePolicyAndExternalData() (*language.PolicyNamespace, *external.Data) {
	// pre-generate the list of labels
	gen.makeLabels()
	fmt.Printf("Labels: %d\n", len(gen.generatedLabels))

	// generate services
	maxChainLen := gen.makeServices()
	fmt.Printf("Services: %d (max chain length %d)\n", len(gen.policy.Services), maxChainLen)

	// generate contexts
	gen.makeContexts()
	fmt.Printf("Contexts: %d\n", len(gen.policy.Contexts))

	// generate dependencies
	gen.makeDependencies()
	fmt.Printf("Dependencies: %d\n", len(gen.policy.Dependencies.DependenciesByID))

	// every user will have the same set of labels
	gen.externalData = external.NewData(
		NewUserLoaderImpl(gen.users, gen.generatedLabels),
		NewSecretLoaderImpl(),
	)

	// there will be one context matching for each service. it will re-define some of those labels
	// there will be other contexts, not matching
	return gen.policy, gen.externalData
}

func (gen *PolicyGenerator) randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		if i == 0 {
			// first letter non-numeric
			b[i] = charset[gen.random.Intn(len(charset)-10)]
		} else {
			// other letters any
			b[i] = charset[gen.random.Intn(len(charset))]
		}
	}
	return string(b)
}

func (gen *PolicyGenerator) makeLabels() {
	gen.generatedLabels = make(map[string]string)
	for i := 0; i < gen.labels; i++ {
		name := gen.randomString(10)
		value := gen.randomString(25)
		gen.generatedLabels[name] = value
	}

	gen.generatedLabelKeys = []string{}
	for key, _ := range gen.generatedLabels {
		gen.generatedLabelKeys = append(gen.generatedLabelKeys, key)
	}
}

func (gen *PolicyGenerator) makeServices() int {
	gen.generatedServices = make([]*language.Service, gen.services)
	for i := 0; i < gen.services; i++ {
		gen.generatedServices[i] = gen.makeService()
	}

	// add some dependencies
	cnt := make([]int, gen.services)
	maxChainLen := 0
	for i := 0; i < gen.services; i++ {
		if maxChainLen < cnt[i] {
			maxChainLen = cnt[i]
		}

		// see if we have exceeded the max chain length
		if cnt[i]+1 > gen.serviceDependencyMaxChain {
			continue
		}

		// try to add at most one dependency from each service
		j := gen.random.Intn(gen.services)
		if j <= i {
			continue
		}

		if cnt[j] < cnt[i]+1 {
			cnt[j] = cnt[i] + 1
		}

		component := &language.ServiceComponent{
			Name:    "dep-" + strconv.Itoa(i),
			Service: "service-" + strconv.Itoa(j),
		}
		gen.generatedServices[i].Components = append(gen.generatedServices[i].Components, component)
	}
	return maxChainLen
}

func (gen *PolicyGenerator) makeService() *language.Service {
	id := len(gen.policy.Services)

	service := &language.Service{
		Metadata: object.Metadata{
			Kind:      language.ServiceObject.Kind,
			Namespace: "main",
			Name:      "service-" + strconv.Itoa(id),
		},
		Owner:      "user-" + strconv.Itoa(gen.random.Intn(gen.users)),
		Components: []*language.ServiceComponent{},
	}

	for i := 0; i < gen.serviceCodeComponents; i++ {
		labelName := gen.generatedLabelKeys[gen.random.Intn(len(gen.generatedLabelKeys))]

		params := util.NestedParameterMap{}
		for j := 0; j < gen.codeParams; j++ {
			name := "param-" + strconv.Itoa(j)
			value := "prefix-{{ .Labels." + labelName + " }}-suffix"
			params[name] = value
		}

		component := &language.ServiceComponent{
			Name: "component-" + strconv.Itoa(i),
			Code: &language.Code{
				Type:   "aptomi/code/unittests",
				Params: params,
			},
		}
		service.Components = append(service.Components, component)
	}

	gen.policy.AddObject(service)
	return service
}

func (gen *PolicyGenerator) makeContexts() {
	for i := 0; i < gen.services; i++ {
		// generate non-matching contexts
		for i := 0; i < gen.contextsPerService-1; i++ {
			context := &language.Context{
				Metadata: object.Metadata{
					Kind:      language.ContextObject.Kind,
					Namespace: "main",
					Name:      "context-" + gen.randomString(20),
				},
				Criteria: &language.Criteria{
					RequireAll: []string{"true"},
					RequireAny: []string{
						gen.randomString(20) + "=='" + gen.randomString(20) + "'",
						gen.randomString(20) + "=='" + gen.randomString(20) + "'",
						gen.randomString(20) + "=='" + gen.randomString(20) + "'",
					},
				},
			}
			gen.policy.AddObject(context)
		}

		// generate matching context
		context := &language.Context{
			Metadata: object.Metadata{
				Kind:      language.ContextObject.Kind,
				Namespace: "main",
				Name:      "context-" + gen.randomString(20),
			},
			Criteria: &language.Criteria{
				RequireAll: []string{"true"},
				RequireAny: []string{"service.Name == '" + ("service-" + strconv.Itoa(i)) + "'"},
			},
		}
		gen.policy.AddObject(context)
	}
}

func (gen *PolicyGenerator) makeDependencies() {
	for i := 0; i < gen.dependencies; i++ {
		dependency := &language.Dependency{
			Metadata: object.Metadata{
				Kind:      language.DependencyObject.Kind,
				Namespace: "main",
				Name:      "dependency-" + strconv.Itoa(i),
			},
			UserID:  "user-" + strconv.Itoa(gen.random.Intn(gen.users)),
			Service: "service-" + strconv.Itoa(gen.random.Intn(gen.services)),
		}
		gen.policy.AddObject(dependency)
	}
}

type UserLoaderImpl struct {
	users  int
	labels map[string]string

	cachedUsers *language.GlobalUsers
}

func NewUserLoaderImpl(users int, labels map[string]string) *UserLoaderImpl {
	return &UserLoaderImpl{
		users:  users,
		labels: labels,
	}
}

func (loader *UserLoaderImpl) LoadUsersAll() language.GlobalUsers {
	if loader.cachedUsers == nil {
		userMap := make(map[string]*language.User)
		for i := 0; i < loader.users; i++ {
			user := &language.User{
				ID:     "user-" + strconv.Itoa(i),
				Name:   "user-" + strconv.Itoa(i),
				Labels: loader.labels,
			}
			userMap[user.ID] = user
		}
		loader.cachedUsers = &language.GlobalUsers{Users: userMap}
	}
	return *loader.cachedUsers
}

func (loader *UserLoaderImpl) LoadUserByID(id string) *language.User {
	return loader.LoadUsersAll().Users[id]
}

func (loader *UserLoaderImpl) Summary() string {
	return "Synthetic user loader"
}

type SecretLoaderImpl struct {
}

func NewSecretLoaderImpl() *SecretLoaderImpl {
	return &SecretLoaderImpl{}
}

func (loader *SecretLoaderImpl) LoadSecretsByUserID(string) language.LabelSet {
	return language.NewLabelSetEmpty()
}

func RunEngine(t *testing.T, desiredPolicy *language.PolicyNamespace, externalData *external.Data) {
	timeStart := time.Now()

	actualPolicy := language.NewPolicyNamespace()
	actualState := resolvePolicy(t, actualPolicy, externalData)
	desiredState := resolvePolicy(t, desiredPolicy, externalData)

	// make plugin to successfully process all components, while failing all instances of component2
	pluginApplyFailComponent2 := NewEnginePluginImpl([]string{"fail-components-like-these"})

	// process all actions
	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).Actions
	plugins := []plugin.EnginePlugin{pluginApplyFailComponent2}
	apply := NewEngineApply(
		desiredPolicy,
		desiredState,
		actualPolicy,
		actualState,
		externalData,
		actions,
		plugins,
	)

	actualState = applyAndCheck(t, apply, ResSuccess, 0, "")

	timeEnd := time.Now()
	timeDiff := timeEnd.Sub(timeStart)

	dResolved := 0
	for _, d := range desiredPolicy.Dependencies.DependenciesByID {
		if d.Resolved {
			dResolved++
		}
	}

	fmt.Printf("Resolved (dependencies): %d\n", dResolved)
	fmt.Printf("Resolved (components): %d\n", len(actualState.ComponentInstanceMap))
	fmt.Printf("Time: %s\n", timeDiff.String())
}
