// nolint: goconst
package diff

import (
	"fmt"
	"testing"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDiffEmpty(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)
	resolvedNext := resolvePolicy(t, b)

	// diff should be empty
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 0, 0, 0, 0, 0)
}

func TestDiffComponentCreationAndAttachClaim(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add claim
	c1 := b.AddClaim(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	c1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 2, 0, 0, 2, 0)

	// add another claim
	c2 := b.AddClaim(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	c2.Labels["param"] = "value1"
	resolvedNextAgain := resolvePolicy(t, b)

	// component should not be instantiated again (it's already there), just new claim should be attached
	diffAgain := NewPolicyResolutionDiff(resolvedNextAgain, resolvedNext)
	verifyDiff(t, diffAgain, 0, 0, 0, 2, 0)
}

func TestDiffComponentUpdate(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add claim
	c1 := b.AddClaim(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	c1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 2, 0, 0, 2, 0)

	// update claim
	c1.Labels["param"] = "value2"
	resolvedNextAgain := resolvePolicy(t, b)

	// component should be updated
	diffAgain := NewPolicyResolutionDiff(resolvedNextAgain, resolvedNext)
	verifyDiff(t, diffAgain, 0, 0, 2, 0, 0)
}

func TestDiffComponentDelete(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add claim
	c1 := b.AddClaim(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	c1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 2, 0, 0, 2, 0)

	// resolve empty policy
	resolvedEmpty := resolvePolicy(t, builder.NewPolicyBuilder())

	// diff should contain destructed component
	diffAgain := NewPolicyResolutionDiff(resolvedEmpty, resolvedNext)
	verifyDiff(t, diffAgain, 0, 2, 0, 0, 2)
}

func TestDiffComponentWithServiceSharing(t *testing.T) {
	b := makePolicyBuilderWithServiceSharing()
	resolvedNext := resolvePolicy(t, b)
	resolvedEmpty := resolvePolicy(t, builder.NewPolicyBuilder())

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedEmpty)
	verifyDiff(t, diff, 7, 0, 0, 9, 0)
}

/*
	Helpers
*/

func makePolicyBuilder() *builder.PolicyBuilder {
	b := builder.NewPolicyBuilder()

	// create a service
	service := b.AddService()
	b.AddServiceComponent(service,
		b.CodeComponent(
			util.NestedParameterMap{"param": "{{ .Labels.param }}"},
			nil,
		),
	)
	b.AddContract(service, b.CriteriaTrue())

	// add rule to set cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, clusterObj.Name)))

	return b
}

func makePolicyBuilderWithServiceSharing() *builder.PolicyBuilder {
	b := builder.NewPolicyBuilder()

	// create a service, which depends on another service
	service1 := b.AddService()
	contract1 := b.AddContract(service1, b.CriteriaTrue())
	service2 := b.AddService()
	contract2 := b.AddContract(service2, b.CriteriaTrue())
	b.AddServiceComponent(service1, b.ContractComponent(contract2))

	// make first service one per claim, and they all will share the second service
	contract1.Contexts[0].Allocation.Keys = []string{"{{ .Claim.ID }}"}

	// add rule to set cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelTarget, clusterObj.Name)))

	// add claims
	b.AddClaim(b.AddUser(), contract1)
	b.AddClaim(b.AddUser(), contract1)
	b.AddClaim(b.AddUser(), contract1)

	return b
}

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder) *resolve.PolicyResolution {
	t.Helper()
	eventLog := event.NewLog(logrus.DebugLevel, "test-resolve")
	resolver := resolve.NewPolicyResolver(builder.Policy(), builder.External(), eventLog)
	result := resolver.ResolveAllClaims()

	claims := builder.Policy().GetObjectsByKind(lang.ClaimObject.Kind)
	for _, claim := range claims {
		if !assert.True(t, result.GetClaimResolution(claim.(*lang.Claim)).Resolved, "Claim resolution status should be correct for %v", claim) {
			hook := event.NewHookConsole(logrus.DebugLevel)
			eventLog.Save(hook)
			t.FailNow()
		}
	}
	return result
}

func verifyDiff(t *testing.T, diff *PolicyResolutionDiff, componentInstantiate int, componentDestruct int, componentUpdate int, componentAttachClaim int, componentDetachClaim int) {
	t.Helper()
	cnt := struct {
		create    int
		update    int
		delete    int
		attach    int
		detach    int
		endpoints int
	}{}

	s := []string{}
	fn := func(act action.Interface) error {
		switch act.(type) {
		case *component.CreateAction:
			cnt.create++
		case *component.DeleteAction:
			cnt.delete++
		case *component.UpdateAction:
			cnt.update++
		case *component.AttachClaimAction:
			cnt.attach++
		case *component.DetachClaimAction:
			cnt.detach++
		case *component.EndpointsAction:
			cnt.endpoints++
		default:
			t.Fatalf("Incorrect action type: %T", act)
		}
		s = append(s, fmt.Sprintf("\n%+v", act))
		return nil
	}

	_ = diff.ActionPlan.Apply(action.WrapSequential(fn), action.NewApplyResultUpdaterImpl())

	ok := assert.Equal(t, componentInstantiate, cnt.create, "Diff: component instantiations")
	ok = ok && assert.Equal(t, componentDestruct, cnt.delete, "Diff: component destructions")
	ok = ok && assert.Equal(t, componentUpdate, cnt.update, "Diff: component updates")
	ok = ok && assert.Equal(t, componentAttachClaim, cnt.attach, "Diff: claims attached to components")
	ok = ok && assert.Equal(t, componentDetachClaim, cnt.detach, "Diff: claims removed from components")
	ok = ok && assert.Equal(t, 0, cnt.endpoints, "Diff: component endpoint actions should never be generated here")

	if !ok {
		t.Logf("Log of actions: %s", s)
		t.FailNow()
	}
}
