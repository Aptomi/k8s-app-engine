package diff

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiffEmpty(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)
	resolvedNext := resolvePolicy(t, b)

	// diff should be empty
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 0, 0, 0, 0, 0, 0)
}

func TestDiffComponentCreationAndAttachDependency(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add dependency
	d1 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 2, 0, 0, 2, 0, 1)

	// add another dependency
	d2 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d2.Labels["param"] = "value1"
	resolvedNextAgain := resolvePolicy(t, b)

	// component should not be instantiated again (it's already there), just new dependency should be attached
	diffAgain := NewPolicyResolutionDiff(resolvedNextAgain, resolvedNext)
	verifyDiff(t, diffAgain, 0, 0, 0, 2, 0, 1)
}

func TestDiffComponentUpdate(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add dependency
	d1 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 2, 0, 0, 2, 0, 1)

	// update dependency
	d1.Labels["param"] = "value2"
	resolvedNextAgain := resolvePolicy(t, b)

	// component should be updated
	diffAgain := NewPolicyResolutionDiff(resolvedNextAgain, resolvedNext)
	verifyDiff(t, diffAgain, 0, 0, 2, 0, 0, 1)
}

func TestDiffComponentDelete(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add dependency
	d1 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev)
	verifyDiff(t, diff, 2, 0, 0, 2, 0, 1)

	// resolve empty policy
	resolvedEmpty := resolvePolicy(t, builder.NewPolicyBuilder())

	// diff should contain destructed component
	diffAgain := NewPolicyResolutionDiff(resolvedEmpty, resolvedNext)
	verifyDiff(t, diffAgain, 0, 2, 0, 0, 2, 0)
}

func TestDiffComponentWithServiceSharing(t *testing.T) {
	b := makePolicyBuilderWithServiceSharing()
	resolvedNext := resolvePolicy(t, b)
	resolvedEmpty := resolvePolicy(t, builder.NewPolicyBuilder())

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedEmpty)
	verifyDiff(t, diff, 7, 0, 0, 9, 0, 0)
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
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, clusterObj.Name)))

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

	// make first service one per dependency, and they all will share the second service
	contract1.Contexts[0].Allocation.Keys = []string{"{{ .Dependency.ID }}"}

	// add rule to set cluster
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, clusterObj.Name)))

	// add dependencies
	b.AddDependency(b.AddUser(), contract1)
	b.AddDependency(b.AddUser(), contract1)
	b.AddDependency(b.AddUser(), contract1)

	return b
}

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder) *resolve.PolicyResolution {
	t.Helper()
	eventLog := event.NewLog("test-resolve", false)
	resolver := resolve.NewPolicyResolver(builder.Policy(), builder.External(), eventLog)
	result := resolver.ResolveAllDependencies()
	if !assert.True(t, result.AllDependenciesResolvedSuccessfully(), "All dependencies should be resolved successfully") {
		hook := &event.HookConsole{}
		eventLog.Save(hook)
		t.FailNow()
	}
	return result
}

func verifyDiff(t *testing.T, diff *PolicyResolutionDiff, componentInstantiate int, componentDestruct int, componentUpdate int, componentAttachDependency int, componentDetachDependency int, componentEndpoints int) {
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
	fn := action.WrapSequential(func(act action.Base) error {
		switch act.(type) {
		case *component.CreateAction:
			cnt.create++
		case *component.DeleteAction:
			cnt.delete++
		case *component.UpdateAction:
			cnt.update++
		case *component.AttachDependencyAction:
			cnt.attach++
		case *component.DetachDependencyAction:
			cnt.detach++
		case *component.EndpointsAction:
			cnt.endpoints++
		default:
			t.Fatalf("Incorrect action type: %T", act)
		}
		s = append(s, fmt.Sprintf("\n%+v", act))
		return nil
	})

	_ = diff.ActionPlan.Apply(fn)

	ok := assert.Equal(t, componentInstantiate, cnt.create, "Diff: component instantiations")
	ok = ok && assert.Equal(t, componentDestruct, cnt.delete, "Diff: component destructions")
	ok = ok && assert.Equal(t, componentUpdate, cnt.update, "Diff: component updates")
	ok = ok && assert.Equal(t, componentAttachDependency, cnt.attach, "Diff: dependencies attached to components")
	ok = ok && assert.Equal(t, componentDetachDependency, cnt.detach, "Diff: dependencies removed from components")
	ok = ok && assert.Equal(t, componentEndpoints, cnt.endpoints, "Diff: component endpoints")

	if !ok {
		t.Logf("Log of actions: %s", s)
		t.FailNow()
	}
}
