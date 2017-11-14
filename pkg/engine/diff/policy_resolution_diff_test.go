package diff

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/global"
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
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev, 0)
	verifyDiff(t, diff, 0, 0, 0, 0, 0, 0, 0)
}

func TestDiffComponentCreationAndAttachDependency(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add dependency
	d1 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev, 0)
	verifyDiff(t, diff, 2, 0, 0, 2, 0, 2, 1)

	// add another dependency
	d2 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d2.Labels["param"] = "value1"
	resolvedNextAgain := resolvePolicy(t, b)

	// component should not be instantiated again (it's already there), just new dependency should be attached
	diffAgain := NewPolicyResolutionDiff(resolvedNextAgain, resolvedNext, 0)
	verifyDiff(t, diffAgain, 0, 0, 0, 2, 0, 2, 1)
}

func TestDiffComponentUpdate(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add dependency
	d1 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev, 0)
	verifyDiff(t, diff, 2, 0, 0, 2, 0, 2, 1)

	// update dependency
	d1.Labels["param"] = "value2"
	resolvedNextAgain := resolvePolicy(t, b)

	// component should be updated
	diffAgain := NewPolicyResolutionDiff(resolvedNextAgain, resolvedNext, 0)
	verifyDiff(t, diffAgain, 0, 0, 2, 0, 0, 1, 1)
}

func TestDiffComponentDelete(t *testing.T) {
	b := makePolicyBuilder()
	resolvedPrev := resolvePolicy(t, b)

	// add dependency
	d1 := b.AddDependency(b.AddUser(), b.Policy().GetObjectsByKind(lang.ContractObject.Kind)[0].(*lang.Contract))
	d1.Labels["param"] = "value1"
	resolvedNext := resolvePolicy(t, b)

	// diff should contain instantiated component
	diff := NewPolicyResolutionDiff(resolvedNext, resolvedPrev, 0)
	verifyDiff(t, diff, 2, 0, 0, 2, 0, 2, 1)

	// resolve empty policy
	resolvedEmpty := resolvePolicy(t, builder.NewPolicyBuilder())

	// diff should contain destructed component
	diffAgain := NewPolicyResolutionDiff(resolvedEmpty, resolvedNext, 0)
	verifyDiff(t, diffAgain, 0, 2, 0, 0, 2, 2, 1)
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

	// add rules to allow all dependencies
	clusterObj := b.AddCluster()
	b.AddRule(b.CriteriaTrue(), b.RuleActions(lang.NewLabelOperationsSetSingleLabel(lang.LabelCluster, clusterObj.Name)))

	return b
}

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder) *resolve.PolicyResolution {
	t.Helper()
	resolver := resolve.NewPolicyResolver(builder.Policy(), builder.External())
	result, eventLog, err := resolver.ResolveAllDependencies()
	if !assert.NoError(t, err, "Policy should be resolved without errors") {
		hook := &event.HookConsole{}
		eventLog.Save(hook)
		t.FailNow()
	}
	return result
}

func verifyDiff(t *testing.T, diff *PolicyResolutionDiff, componentInstantiate int, componentDestruct int, componentUpdate int, componentAttachDependency int, componentDetachDependency int, componentEndpoints int, clusters int) {
	t.Helper()
	cnt := struct {
		create    int
		update    int
		delete    int
		attach    int
		detach    int
		endpoints int
		clusters  int
	}{}
	s := []string{}
	for _, act := range diff.Actions {
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
		case *global.PostProcessAction:
			cnt.clusters++
		default:
			t.Fatalf("Incorrect action type: %T", act)
		}
		s = append(s, fmt.Sprintf("%+v", act))
	}

	ok := assert.Equal(t, componentInstantiate, cnt.create, "Diff: component instantiations")
	ok = ok && assert.Equal(t, componentDestruct, cnt.delete, "Diff: component destructions")
	ok = ok && assert.Equal(t, componentUpdate, cnt.update, "Diff: component updates")
	ok = ok && assert.Equal(t, componentAttachDependency, cnt.attach, "Diff: dependencies attached to components")
	ok = ok && assert.Equal(t, componentDetachDependency, cnt.detach, "Diff: dependencies removed from components")
	ok = ok && assert.Equal(t, componentEndpoints, cnt.endpoints, "Diff: component endpoints")
	ok = ok && assert.Equal(t, clusters, cnt.clusters, "Diff: all clusters post processing")

	if !ok {
		t.Logf("Log of diff actions: %s", s)
		t.FailNow()
	}
}
