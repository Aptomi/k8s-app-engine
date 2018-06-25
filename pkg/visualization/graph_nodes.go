package visualization

import (
	"fmt"
	"html"

	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

/*
	Claim
*/
type claimNode struct {
	claim *lang.Claim
	b     *GraphBuilder
}

func (n claimNode) getGroup() string {
	dResolution := n.b.resolution.GetClaimResolution(n.claim)
	if dResolution.Resolved {
		return "claim"
	}
	return "claimNotResolved"
}

func (n claimNode) getID() string {
	return runtime.KeyForStorable(n.claim)
}

func (n claimNode) getLabel() string {
	return n.claim.Metadata.Namespace + "/" + n.claim.Name
}

/*
	Contract
*/
type contractNode struct {
	contract *lang.Contract
}

func (n contractNode) getGroup() string {
	return "contract"
}

func (n contractNode) getID() string {
	return runtime.KeyForStorable(n.contract)
}

func (n contractNode) getLabel() string {
	return fmt.Sprintf(
		`contract: <b>%s</b>`,
		html.EscapeString(n.contract.Name),
	)
}

/*
	Bundle
*/
type bundleNode struct {
	bundle *lang.Bundle
}

func (n bundleNode) getGroup() string {
	return "bundle"
}

func (n bundleNode) getID() string {
	return runtime.KeyForStorable(n.bundle)
}

func (n bundleNode) getLabel() string {
	return n.bundle.Name
}

/*
	Component
*/
type componentNode struct {
	bundle    *lang.Bundle
	component *lang.BundleComponent
}

func (n componentNode) getGroup() string {
	return "component" + n.component.Code.Type
}

func (n componentNode) getID() string {
	return runtime.KeyForStorable(n.bundle) + "-" + n.component.Name
}

func (n componentNode) getLabel() string {
	return n.component.Name
}

/*
	Bundle Instance
*/
type bundleInstanceNode struct {
	instance *resolve.ComponentInstance
	bundle   *lang.Bundle
}

func (n bundleInstanceNode) getGroup() string {
	return "bundleInstance"
}

func (n bundleInstanceNode) getID() string {
	return n.instance.GetKey()
}

func (n bundleInstanceNode) getLabel() string {
	result := fmt.Sprintf(
		`<b>%s</b>
				context: <i>%s</i>`,
		html.EscapeString(n.bundle.Name),
		html.EscapeString(n.instance.Metadata.Key.ContextName),
	)

	if len(n.instance.Metadata.Key.KeysResolved) > 0 {
		result += fmt.Sprintf("\nkeys: <i>%s</i>", html.EscapeString(shorten(n.instance.Metadata.Key.KeysResolved)))
	}

	result += fmt.Sprintf("\ntarget: <i>%s</i>", html.EscapeString(getTargetName(n.instance.Metadata.Key)))

	if !n.instance.CreatedAt.IsZero() {
		result += fmt.Sprintf("\nrunning: <i>%s</i>", html.EscapeString(n.instance.GetRunningTime().String()))
	}

	return result
}

func getTargetName(key *resolve.ComponentInstanceKey) string {
	result := key.ClusterName
	if key.ClusterNameSpace != runtime.SystemNS {
		result = key.ClusterNameSpace + "/" + result
	}
	if len(key.TargetSuffix) > 0 {
		result += "." + key.TargetSuffix
	}
	return result
}

func shorten(s string) string {
	if len(s) > 15 {
		suffix := "..."
		return s[:15-len(suffix)] + suffix
	}
	return s
}

/*
	Error
*/
type errorNode struct {
	err error
}

func (n errorNode) getGroup() string {
	return "error"
}

func (n errorNode) getID() string {
	return fmt.Sprintf("error-%p", n.err) // nolint: vet
}

func (n errorNode) getLabel() string {
	return fmt.Sprintf("Error: %s", n.err)
}
