package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkEngineSmall(b *testing.B) {
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
		RunEngine(b, "small", smallPolicy, smallExternalData)
	}
}

func BenchmarkEngineMedium(b *testing.B) {
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
		RunEngine(b, "medium", mediumPolicy, mediumExternalData)
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

	err := gen.policy.Validate()
	if err != nil {
		panic(err)
	}
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
		TypeKind: lang.ServiceObject.GetTypeKind(),
		Metadata: lang.Metadata{
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
				Type:   "helm",
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
			TypeKind: lang.RuleObject.GetTypeKind(),
			Metadata: lang.Metadata{
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
		TypeKind: lang.RuleObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: "main",
			Name:      "rule-" + strconv.Itoa(gen.rules),
		},
		Weight: gen.rules,
		Criteria: &lang.Criteria{
			RequireAll: []string{"true"},
		},
		Actions: &lang.RuleActions{
			ChangeLabels: lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, "cluster-test"),
		},
	})
}

func (gen *PolicyGenerator) makeContracts() {
	for i := 0; i < gen.services; i++ {
		contract := &lang.Contract{
			TypeKind: lang.ContractObject.GetTypeKind(),
			Metadata: lang.Metadata{
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
			TypeKind: lang.DependencyObject.GetTypeKind(),
			Metadata: lang.Metadata{
				Namespace: "main",
				Name:      "dependency-" + strconv.Itoa(i),
			},
			User:     "user-" + strconv.Itoa(gen.random.Intn(gen.users)),
			Contract: "contract-" + strconv.Itoa(gen.random.Intn(gen.services)),
		}
		gen.addObject(dependency)
	}
}

func (gen *PolicyGenerator) makeCluster() {
	cluster := &lang.Cluster{
		TypeKind: lang.ClusterObject.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: "system",
			Name:      "cluster-test",
		},
		Type: "kubernetes",
		Config: struct {
			Namespace string
		}{
			Namespace: "default",
		},
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
				Name:        "user-" + strconv.Itoa(i),
				Labels:      loader.labels,
				DomainAdmin: true,
			}
			userMap[user.Name] = user
		}
		loader.cachedUsers = &lang.GlobalUsers{Users: userMap}
	}
	return loader.cachedUsers
}

func (loader *UserLoaderImpl) LoadUserByName(id string) *lang.User {
	return loader.LoadUsersAll().Users[id]
}

func (loader *UserLoaderImpl) Authenticate(userName, password string) (*lang.User, error) {
	return nil, nil
}

func (loader *UserLoaderImpl) Summary() string {
	return "Synthetic user loader"
}

type SecretLoaderImpl struct {
}

func NewSecretLoaderImpl() *SecretLoaderImpl {
	return &SecretLoaderImpl{}
}

func (loader *SecretLoaderImpl) LoadSecretsByUserName(string) map[string]string {
	return nil
}

func RunEngine(b *testing.B, testName string, desiredPolicy *lang.Policy, externalData *external.Data) {
	fmt.Printf("Running engine for '%s'\n", testName)

	timeStart := time.Now()

	// resolve all dependencies and apply actions
	actualState := resolvePolicyBenchmark(b, lang.NewPolicy(), externalData, false)
	desiredState := resolvePolicyBenchmark(b, desiredPolicy, externalData, true)

	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan
	applier := NewEngineApply(
		desiredPolicy,
		desiredState,
		actualState,
		actual.NewNoOpActionStateUpdater(),
		externalData,
		mockRegistry(true, false),
		actions,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)
	actualState = applyAndCheckBenchmark(b, applier, action.ApplyResult{Success: applier.actionPlan.NumberOfActions(), Failed: 0, Skipped: 0})

	timeCheckpoint := time.Now()
	fmt.Printf("[%s] Time = %s, resolving %d dependencies and %d component instances\n", testName, time.Since(timeStart).String(), desiredState.SuccessfullyResolvedDependencies(), len(actualState.ComponentInstanceMap))

	// now, remove all dependencies and apply actions
	for _, dependency := range desiredPolicy.GetObjectsByKind(lang.DependencyObject.Kind) {
		desiredPolicy.RemoveObject(dependency)
	}
	desiredState = resolvePolicyBenchmark(b, desiredPolicy, externalData, false)
	actions = diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan
	applier = NewEngineApply(
		desiredPolicy,
		desiredState,
		actualState,
		actual.NewNoOpActionStateUpdater(),
		externalData,
		mockRegistry(true, false),
		actions,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)
	_ = applyAndCheckBenchmark(b, applier, action.ApplyResult{Success: applier.actionPlan.NumberOfActions(), Failed: 0, Skipped: 0})

	fmt.Printf("[%s] Time = %s, deleting all dependencies and component instances\n", testName, time.Since(timeCheckpoint).String())
}

func applyAndCheckBenchmark(b *testing.B, apply *EngineApply, expectedResult action.ApplyResult) *resolve.PolicyResolution {
	b.Helper()
	actualState, result := apply.Apply()

	t := &testing.T{}
	ok := assert.Equal(t, expectedResult.Success, result.Success, "Number of successfully executed actions")
	ok = ok && assert.Equal(t, expectedResult.Failed, result.Failed, "Number of failed actions")
	ok = ok && assert.Equal(t, expectedResult.Skipped, result.Skipped, "Number of skipped actions")
	ok = ok && assert.Equal(t, expectedResult.Success+expectedResult.Failed+expectedResult.Skipped, result.Total, "Number of total actions")

	if !ok {
		// print log into stdout and exit
		hook := event.NewHookConsole(logrus.DebugLevel)
		apply.eventLog.Save(hook)
		b.Fatal("action counts don't match")
	}

	return actualState
}

func resolvePolicyBenchmark(b *testing.B, policy *lang.Policy, externalData *external.Data, expectedNonEmpty bool) *resolve.PolicyResolution {
	b.Helper()
	eventLog := event.NewLog(logrus.DebugLevel, "test-resolve")
	resolver := resolve.NewPolicyResolver(policy, externalData, eventLog)
	result := resolver.ResolveAllDependencies()
	t := &testing.T{}
	if !assert.True(t, result.AllDependenciesResolvedSuccessfully(), "All dependencies should be resolved successfully") {
		hook := event.NewHookConsole(logrus.DebugLevel)
		eventLog.Save(hook)
		b.Fatal("policy resolution error")
	}

	if expectedNonEmpty != (result.SuccessfullyResolvedDependencies() > 0) {
		hook := event.NewHookConsole(logrus.DebugLevel)
		eventLog.Save(hook)
		b.Fatal("no dependencies resolved")
	}

	return result
}

func (gen *PolicyGenerator) addObject(obj lang.Base) {
	err := gen.policy.AddObject(obj)
	if err != nil {
		panic(err)
	}
}
