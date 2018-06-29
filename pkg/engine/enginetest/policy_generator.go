package enginetest

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
)

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

func (gen *PolicyGenerator) MakePolicyAndExternalData() (*lang.Policy, *external.Data) {
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

func (gen *PolicyGenerator) addObject(obj lang.Base) {
	err := gen.policy.AddObject(obj)
	if err != nil {
		panic(err)
	}
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
