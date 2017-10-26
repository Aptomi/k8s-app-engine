package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkEngineSmall(b *testing.B) {
	t := &testing.T{}

	// small policy
	smallPolicy, smallExternalData := NewPolicyGenerator(
		239,
		30,
		50,
		6,
		6,
		4,
		2,
		25,
		100,
		500,
	).makePolicyAndExternalData()

	for i := 0; i < b.N; i++ {
		RunEngine(t, "small", smallPolicy, smallExternalData)
	}
}

func BenchmarkEngineMedium(b *testing.B) {
	t := &testing.T{}

	// medium policy
	mediumPolicy, mediumExternalData := NewPolicyGenerator(
		239,
		30,
		100,
		6,
		6,
		4,
		2,
		25,
		10000,
		2000,
	).makePolicyAndExternalData()

	for i := 0; i < b.N; i++ {
		RunEngine(t, "medium", mediumPolicy, mediumExternalData)
	}
}

type PolicyGenerator struct {
	random                    *rand.Rand
	labels                    int
	services                  int
	serviceCodeComponents     int
	codeParams                int
	serviceDependencyMaxChain int
	contextsPerContract       int
	rules                     int
	users                     int
	dependencies              int

	generatedUserLabels map[string]string
	generatedLabelKeys  []string

	generatedServices []*lang.Service
	policy            *lang.Policy
	externalData      *external.Data
}

func NewPolicyGenerator(randSeed int64, labels, services, serviceCodeComponents, codeParams, serviceDependencyMaxChain, contextsPerContract, rules, users, dependencies int) *PolicyGenerator {
	result := &PolicyGenerator{
		random:                    rand.New(rand.NewSource(randSeed)),
		labels:                    labels,
		services:                  services,
		serviceCodeComponents:     serviceCodeComponents,
		codeParams:                codeParams,
		serviceDependencyMaxChain: serviceDependencyMaxChain,
		contextsPerContract:       contextsPerContract,
		rules:                     rules,
		users:                     users,
		dependencies:              dependencies,
		policy:                    lang.NewPolicy(),
	}
	return result
}

func (gen *PolicyGenerator) makePolicyAndExternalData() (*lang.Policy, *external.Data) {
	// pre-generate the list of labels
	gen.makeUserLabels()

	// generate services
	maxChainLen := gen.makeServices()

	// generate contracts
	gen.makeContracts()

	// generate rules
	gen.makeRules()

	// generate dependencies
	gen.makeDependencies()

	// generate cluster
	gen.makeCluster()

	// every user will have the same set of labels
	gen.externalData = external.NewData(
		NewUserLoaderImpl(gen.users, gen.generatedUserLabels),
		NewSecretLoaderImpl(),
	)

	fmt.Printf("Generated policy. Services = %d (max chain %d), Contexts = %d, Dependencies = %d, Users = %d\n",
		len(gen.policy.GetObjectsByKind(lang.ServiceObject.Kind)),
		maxChainLen,
		len(gen.policy.GetObjectsByKind(lang.ContractObject.Kind))*gen.contextsPerContract,
		len(gen.policy.GetObjectsByKind(lang.DependencyObject.Kind)),
		len(gen.externalData.UserLoader.LoadUsersAll().Users),
	)

	// there will be one context matching for each service. it will re-define some of those labels
	// there will be other contexts, not matching
	return gen.policy, gen.externalData
}

func (gen *PolicyGenerator) makeUserLabels() {
	gen.generatedUserLabels = make(map[string]string)
	for i := 0; i < gen.labels; i++ {
		name := util.RandomID(gen.random, 10)
		value := util.RandomID(gen.random, 25)
		gen.generatedUserLabels[name] = value
	}

	gen.generatedLabelKeys = []string{}
	for key := range gen.generatedUserLabels {
		gen.generatedLabelKeys = append(gen.generatedLabelKeys, key)
	}
}

func (gen *PolicyGenerator) makeServices() int {
	gen.generatedServices = make([]*lang.Service, gen.services)
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

		component := &lang.ServiceComponent{
			Name:     "dep-" + strconv.Itoa(i),
			Contract: "contract-" + strconv.Itoa(j),
		}
		gen.generatedServices[i].Components = append(gen.generatedServices[i].Components, component)
	}
	return maxChainLen
}

func (gen *PolicyGenerator) makeService() *lang.Service {
	id := len(gen.policy.GetObjectsByKind(lang.ServiceObject.Kind))

	service := &lang.Service{
		Metadata: lang.Metadata{
			Kind:      lang.ServiceObject.Kind,
			Namespace: "main",
			Name:      "service-" + strconv.Itoa(id),
		},
		Components: []*lang.ServiceComponent{},
	}

	for i := 0; i < gen.serviceCodeComponents; i++ {
		labelName := gen.generatedLabelKeys[gen.random.Intn(len(gen.generatedLabelKeys))]

		params := util.NestedParameterMap{}
		params[lang.LabelCluster] = "cluster-test"
		for j := 0; j < gen.codeParams; j++ {
			name := "param-" + strconv.Itoa(j)
			value := "prefix-{{ .Labels." + labelName + " }}-suffix"
			params[name] = value
		}

		component := &lang.ServiceComponent{
			Name: "component-" + strconv.Itoa(i),
			Code: &lang.Code{
				Type:   "aptomi/code/unittests",
				Params: params,
			},
		}
		service.Components = append(service.Components, component)
	}

	gen.addObject(service)
	return service
}

func (gen *PolicyGenerator) makeRules() {
	// generate non-matching rules
	for i := 0; i < gen.rules-1; i++ {
		gen.addObject(&lang.Rule{
			Metadata: lang.Metadata{
				Kind:      lang.RuleObject.Kind,
				Namespace: "main",
				Name:      "rule-" + strconv.Itoa(i),
			},
			Weight: i,
			Criteria: &lang.Criteria{
				RequireAll: []string{"service.Name == 'some-name-" + strconv.Itoa(i) + "'"},
			},
			Actions: &lang.RuleActions{
				Dependency: lang.DependencyAction("reject"),
			},
		})
	}

	// generate rule which allows all dependencies
	gen.addObject(&lang.Rule{
		Metadata: lang.Metadata{
			Kind:      lang.RuleObject.Kind,
			Namespace: "main",
			Name:      "rule-" + strconv.Itoa(gen.rules),
		},
		Weight: gen.rules,
		Criteria: &lang.Criteria{
			RequireAll: []string{"true"},
		},
		Actions: &lang.RuleActions{
			ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, "cluster-test")),
		},
	})
}

func (gen *PolicyGenerator) makeContracts() {
	for i := 0; i < gen.services; i++ {
		contract := &lang.Contract{
			Metadata: lang.Metadata{
				Kind:      lang.ContractObject.Kind,
				Namespace: "main",
				Name:      "contract-" + strconv.Itoa(i),
			},
			Contexts: []*lang.Context{},
		}

		// generate non-matching contexts
		for j := 0; j < gen.contextsPerContract-1; j++ {
			context := &lang.Context{
				Name: "context-" + util.RandomID(gen.random, 20),
				Criteria: &lang.Criteria{
					RequireAll: []string{"true"},
					RequireAny: []string{
						util.RandomID(gen.random, 20) + "=='" + util.RandomID(gen.random, 20) + "'",
						util.RandomID(gen.random, 20) + "=='" + util.RandomID(gen.random, 20) + "'",
						util.RandomID(gen.random, 20) + "=='" + util.RandomID(gen.random, 20) + "'",
					},
				},
				Allocation: &lang.Allocation{
					Service: "service-" + strconv.Itoa(i),
				},
			}
			contract.Contexts = append(contract.Contexts, context)
		}

		// generate matching context
		context := &lang.Context{
			Name: "context-" + util.RandomID(gen.random, 20),
			Criteria: &lang.Criteria{
				RequireAll: []string{"true"},
			},
			Allocation: &lang.Allocation{
				Service: "service-" + strconv.Itoa(i),
			},
		}
		contract.Contexts = append(contract.Contexts, context)

		// add service contract to the policy
		gen.addObject(contract)
	}
}

func (gen *PolicyGenerator) makeDependencies() {
	for i := 0; i < gen.dependencies; i++ {
		dependency := &lang.Dependency{
			Metadata: lang.Metadata{
				Kind:      lang.DependencyObject.Kind,
				Namespace: "main",
				Name:      "dependency-" + strconv.Itoa(i),
			},
			UserID:   "user-" + strconv.Itoa(gen.random.Intn(gen.users)),
			Contract: "contract-" + strconv.Itoa(gen.random.Intn(gen.services)),
		}
		gen.addObject(dependency)
	}
}

func (gen *PolicyGenerator) makeCluster() {
	cluster := &lang.Cluster{
		Metadata: lang.Metadata{
			Kind:      lang.ClusterObject.Kind,
			Namespace: "system",
			Name:      "cluster-test",
		},
		Type: "kubernetes",
	}
	gen.addObject(cluster)
}

type UserLoaderImpl struct {
	users  int
	labels map[string]string

	cachedUsers *lang.GlobalUsers
}

func NewUserLoaderImpl(users int, labels map[string]string) *UserLoaderImpl {
	return &UserLoaderImpl{
		users:  users,
		labels: labels,
	}
}

func (loader *UserLoaderImpl) LoadUsersAll() *lang.GlobalUsers {
	if loader.cachedUsers == nil {
		userMap := make(map[string]*lang.User)
		for i := 0; i < loader.users; i++ {
			user := &lang.User{
				ID:     "user-" + strconv.Itoa(i),
				Name:   "user-" + strconv.Itoa(i),
				Labels: loader.labels,
				Admin:  true,
			}
			userMap[user.ID] = user
		}
		loader.cachedUsers = &lang.GlobalUsers{Users: userMap}
	}
	return loader.cachedUsers
}

func (loader *UserLoaderImpl) LoadUserByID(id string) *lang.User {
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

func (loader *SecretLoaderImpl) LoadSecretsByUserID(string) map[string]string {
	return nil
}

func RunEngine(t *testing.T, testName string, desiredPolicy *lang.Policy, externalData *external.Data) {
	fmt.Printf("Running engine for '%s'\n", testName)

	timeStart := time.Now()

	actualPolicy := lang.NewPolicy()
	actualState := resolvePolicyBenchmark(t, actualPolicy, externalData, false)
	desiredState := resolvePolicyBenchmark(t, desiredPolicy, externalData, true)

	// process all actions
	actions := diff.NewPolicyResolutionDiff(desiredState, actualState, 0).Actions

	applier := NewEngineApply(
		desiredPolicy,
		desiredState,
		actualPolicy,
		actualState,
		actual.NewNoOpActionStateUpdater(),
		externalData,
		MockPluginFailOnComponent("fail-components-like-these"),
		actions,
		progress.NewNoop(),
	)

	actualState = applyAndCheck(t, applier, ResSuccess, 0, "Successfully resolved")

	timeEnd := time.Now()
	timeDiff := timeEnd.Sub(timeStart)

	fmt.Printf("[%s] Time = %s, Resolved = dependencies %d, components %d\n", testName, timeDiff.String(), len(desiredState.DependencyInstanceMap), len(actualState.ComponentInstanceMap))
}

func resolvePolicyBenchmark(t *testing.T, policy *lang.Policy, externalData *external.Data, expectedNonEmpty bool) *resolve.PolicyResolution {
	t.Helper()
	resolver := resolve.NewPolicyResolver(policy, externalData)
	result, eventLog, err := resolver.ResolveAllDependencies()
	if !assert.NoError(t, err, "Policy should be resolved without errors") {
		hook := &event.HookConsole{}
		eventLog.Save(hook)
		panic("Policy resolution error")
	}

	if expectedNonEmpty && len(result.DependencyInstanceMap) <= 0 {
		hook := &event.HookConsole{}
		eventLog.Save(hook)
		t.FailNow()
		panic("No dependencies resolved")
	}

	return result
}

func (gen *PolicyGenerator) addObject(obj object.Base) {
	err := gen.policy.AddObject(obj)
	if err != nil {
		panic(err)
	}
}
