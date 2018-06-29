package apply

import (
	"fmt"
	"testing"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/enginetest"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func BenchmarkEngineSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// create small policy
		smallPolicy, smallExternalData := enginetest.NewPolicyGenerator(
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
		).MakePolicyAndExternalData()

		// run tests on small policy
		RunEngine(b, "small", smallPolicy, smallExternalData)
	}
}

func BenchmarkEngineMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// medium policy
		mediumPolicy, mediumExternalData := enginetest.NewPolicyGenerator(
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
		).MakePolicyAndExternalData()

		// run tests on medium policy
		RunEngine(b, "medium", mediumPolicy, mediumExternalData)
	}
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
