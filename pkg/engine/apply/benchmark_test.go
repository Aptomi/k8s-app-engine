package apply

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func BenchmarkEngineSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// create small policy
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

		// run tests on small policy
		RunEngine(b, "small", smallPolicy, smallExternalData)
	}
}

func BenchmarkEngineMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
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

		// run tests on medium policy
		RunEngine(b, "medium", mediumPolicy, mediumExternalData)
	}
}

type PolicyGenerator struct {
	random               *rand.Rand
	labels               int
	bundles              int
	bundleCodeComponents int
	codeParams           int
	bundleClaimMaxChain  int
	contextsPerService   int
	rules                int
	users                int
	claims               int

	generatedUserLabels map[string]string
	generatedLabelKeys  []string

	generatedBundles []*lang.Bundle
	policy           *lang.Policy
	externalData     *external.Data
}

func NewPolicyGenerator(randSeed int64, labels, bundles, bundleCodeComponents, codeParams, bundleClaimMaxChain, contextsPerService, rules, users, claims int) *PolicyGenerator {
	result := &PolicyGenerator{
		random:               rand.New(rand.NewSource(randSeed)),
		labels:               labels,
		bundles:              bundles,
		bundleCodeComponents: bundleCodeComponents,
		codeParams:           codeParams,
		bundleClaimMaxChain:  bundleClaimMaxChain,
		contextsPerService:   contextsPerService,
		rules:                rules,
		users:                users,
		claims:               claims,
		policy:               lang.NewPolicy(),
	}
	return result
}

func (gen *PolicyGenerator) makePolicyAndExternalData() (*lang.Policy, *external.Data) {
	// pre-generate the list of labels
	gen.makeUserLabels()

	// generate bundles
	maxChainLen := gen.makeBundles()

	// generate services
	gen.makeServices()

	// generate rules
	gen.makeRules()

	// generate claims
	gen.makeClaims()

	// generate cluster
	gen.makeCluster()

	// every user will have the same set of labels
	gen.externalData = external.NewData(
		NewUserLoaderImpl(gen.users, gen.generatedUserLabels),
		NewSecretLoaderImpl(),
	)

	fmt.Printf("Generated policy. Bundles = %d (max chain %d), Contexts = %d, Claims = %d, Users = %d\n",
		len(gen.policy.GetObjectsByKind(lang.TypeBundle.Kind)),
		maxChainLen,
		len(gen.policy.GetObjectsByKind(lang.TypeService.Kind))*gen.contextsPerService,
		len(gen.policy.GetObjectsByKind(lang.TypeClaim.Kind)),
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

func (gen *PolicyGenerator) makeBundles() int {
	gen.generatedBundles = make([]*lang.Bundle, gen.bundles)
	for i := 0; i < gen.bundles; i++ {
		gen.generatedBundles[i] = gen.makeBundle()
	}

	// add some dependencies between bundles
	cnt := make([]int, gen.bundles)
	maxChainLen := 0
	for i := 0; i < gen.bundles; i++ {
		if maxChainLen < cnt[i] {
			maxChainLen = cnt[i]
		}

		// see if we have exceeded the max chain length
		if cnt[i]+1 > gen.bundleClaimMaxChain {
			continue
		}

		// try to add at most one claim from each bundle
		j := gen.random.Intn(gen.bundles)
		if j <= i {
			continue
		}

		if cnt[j] < cnt[i]+1 {
			cnt[j] = cnt[i] + 1
		}

		component := &lang.BundleComponent{
			Name:    "dep-" + strconv.Itoa(i),
			Service: "service-" + strconv.Itoa(j),
		}
		gen.generatedBundles[i].Components = append(gen.generatedBundles[i].Components, component)
	}
	return maxChainLen
}

func (gen *PolicyGenerator) makeBundle() *lang.Bundle {
	id := len(gen.policy.GetObjectsByKind(lang.TypeBundle.Kind))

	bundle := &lang.Bundle{
		TypeKind: lang.TypeBundle.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: "main",
			Name:      "bundle-" + strconv.Itoa(id),
		},
		Components: []*lang.BundleComponent{},
	}

	for i := 0; i < gen.bundleCodeComponents; i++ {
		labelName := gen.generatedLabelKeys[gen.random.Intn(len(gen.generatedLabelKeys))]

		params := util.NestedParameterMap{}
		params[lang.LabelTarget] = "cluster-test"
		for j := 0; j < gen.codeParams; j++ {
			name := "param-" + strconv.Itoa(j)
			value := "prefix-{{ .Labels." + labelName + " }}-suffix"
			params[name] = value
		}

		component := &lang.BundleComponent{
			Name: "component-" + strconv.Itoa(i),
			Code: &lang.Code{
				Type:   "helm",
				Params: params,
			},
		}
		bundle.Components = append(bundle.Components, component)
	}

	gen.addObject(bundle)
	return bundle
}

func (gen *PolicyGenerator) makeRules() {
	// generate non-matching rules
	for i := 0; i < gen.rules-1; i++ {
		gen.addObject(&lang.Rule{
			TypeKind: lang.TypeRule.GetTypeKind(),
			Metadata: lang.Metadata{
				Namespace: "main",
				Name:      "rule-" + strconv.Itoa(i),
			},
			Weight: i,
			Criteria: &lang.Criteria{
				RequireAll: []string{"bundle.Name == 'some-name-" + strconv.Itoa(i) + "'"},
			},
			Actions: &lang.RuleActions{
				Claim: lang.ClaimAction("reject"),
			},
		})
	}

	// generate rule which specifies placement for all claims
	gen.addObject(&lang.Rule{
		TypeKind: lang.TypeRule.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: "main",
			Name:      "rule-" + strconv.Itoa(gen.rules),
		},
		Weight: gen.rules,
		Criteria: &lang.Criteria{
			RequireAll: []string{"true"},
		},
		Actions: &lang.RuleActions{
			ChangeLabels: lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, "cluster-test"),
		},
	})
}

func (gen *PolicyGenerator) makeServices() {
	for i := 0; i < gen.bundles; i++ {
		service := &lang.Service{
			TypeKind: lang.TypeService.GetTypeKind(),
			Metadata: lang.Metadata{
				Namespace: "main",
				Name:      "service-" + strconv.Itoa(i),
			},
			Contexts: []*lang.Context{},
		}

		// generate non-matching contexts
		for j := 0; j < gen.contextsPerService-1; j++ {
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
					Bundle: "bundle-" + strconv.Itoa(i),
				},
			}
			service.Contexts = append(service.Contexts, context)
		}

		// generate matching context
		context := &lang.Context{
			Name: "context-" + util.RandomID(gen.random, 20),
			Criteria: &lang.Criteria{
				RequireAll: []string{"true"},
			},
			Allocation: &lang.Allocation{
				Bundle: "bundle-" + strconv.Itoa(i),
			},
		}
		service.Contexts = append(service.Contexts, context)

		// add service to the policy
		gen.addObject(service)
	}
}

func (gen *PolicyGenerator) makeClaims() {
	for i := 0; i < gen.claims; i++ {
		claim := &lang.Claim{
			TypeKind: lang.TypeClaim.GetTypeKind(),
			Metadata: lang.Metadata{
				Namespace: "main",
				Name:      "claim-" + strconv.Itoa(i),
			},
			User:    "user-" + strconv.Itoa(gen.random.Intn(gen.users)),
			Service: "service-" + strconv.Itoa(gen.random.Intn(gen.bundles)),
		}
		gen.addObject(claim)
	}
}

func (gen *PolicyGenerator) makeCluster() {
	cluster := &lang.Cluster{
		TypeKind: lang.TypeCluster.GetTypeKind(),
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

	// resolve all claims and apply actions
	actualState := resolvePolicyBenchmark(b, lang.NewPolicy(), externalData, false)
	desiredState := resolvePolicyBenchmark(b, desiredPolicy, externalData, true)

	actions := diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan
	applier := NewEngineApply(
		desiredPolicy,
		desiredState,
		actual.NewNoOpActionStateUpdater(actualState),
		externalData,
		mockRegistry(true, false),
		actions,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)
	actualState = applyAndCheckBenchmark(b, applier, action.ApplyResult{Success: applier.actionPlan.NumberOfActions(), Failed: 0, Skipped: 0})

	timeCheckpoint := time.Now()
	fmt.Printf("[%s] Time = %s, resolving %d claims and %d component instances\n", testName, time.Since(timeStart).String(), len(desiredPolicy.GetObjectsByKind(lang.TypeClaim.Kind)), len(actualState.ComponentInstanceMap))

	// now, remove all claims and apply actions
	for _, claim := range desiredPolicy.GetObjectsByKind(lang.TypeClaim.Kind) {
		desiredPolicy.RemoveObject(claim)
	}
	desiredState = resolvePolicyBenchmark(b, desiredPolicy, externalData, false)
	actions = diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan
	applier = NewEngineApply(
		desiredPolicy,
		desiredState,
		actual.NewNoOpActionStateUpdater(actualState),
		externalData,
		mockRegistry(true, false),
		actions,
		event.NewLog(logrus.DebugLevel, "test-apply"),
		action.NewApplyResultUpdaterImpl(),
	)
	_ = applyAndCheckBenchmark(b, applier, action.ApplyResult{Success: applier.actionPlan.NumberOfActions(), Failed: 0, Skipped: 0})

	fmt.Printf("[%s] Time = %s, deleting all claims and component instances\n", testName, time.Since(timeCheckpoint))
}

func applyAndCheckBenchmark(b *testing.B, apply *EngineApply, expectedResult action.ApplyResult) *resolve.PolicyResolution {
	b.Helper()
	actualState, result := apply.Apply(50)

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
	result := resolver.ResolveAllClaims()
	t := &testing.T{}

	claims := policy.GetObjectsByKind(lang.TypeClaim.Kind)
	for _, claim := range claims {
		if !assert.True(t, result.GetClaimResolution(claim.(*lang.Claim)).Resolved, "Claim resolution status should be correct for %v", claim) {
			hook := event.NewHookConsole(logrus.DebugLevel)
			eventLog.Save(hook)
			b.Fatal("policy resolution error")
		}
	}

	if expectedNonEmpty != (len(claims) > 0) {
		hook := event.NewHookConsole(logrus.DebugLevel)
		eventLog.Save(hook)
		fmt.Println(expectedNonEmpty)
		fmt.Println(len(claims))
		b.Fatal("no claims resolved")
	}

	return result
}

func (gen *PolicyGenerator) addObject(obj lang.Base) {
	err := gen.policy.AddObject(obj)
	if err != nil {
		panic(err)
	}
}
