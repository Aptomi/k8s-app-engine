package visualization

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"html"
)

/*
	Dependency
*/
type dependencyNode struct {
	dependency *lang.Dependency
	b          *GraphBuilder
}

func (n dependencyNode) getGroup() string {
	dResolution := n.b.resolution.GetDependencyResolution(n.dependency)
	if dResolution.Resolved {
		return "dependency"
	}
	return "dependencyNotResolved"
}

func (n dependencyNode) getID() string {
	return runtime.KeyForStorable(n.dependency)
}

func (n dependencyNode) getLabel() string {
	return n.dependency.Metadata.Namespace + "/" + n.dependency.Name
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
	Service
*/
type serviceNode struct {
	service *lang.Service
}

func (n serviceNode) getGroup() string {
	return "service"
}

func (n serviceNode) getID() string {
	return runtime.KeyForStorable(n.service)
}

func (n serviceNode) getLabel() string {
	return n.service.Name
}

/*
	Component
*/
type componentNode struct {
	service   *lang.Service
	component *lang.ServiceComponent
}

func (n componentNode) getGroup() string {
	return "component" + n.component.Code.Type
}

func (n componentNode) getID() string {
	return runtime.KeyForStorable(n.service) + "-" + n.component.Name
}

func (n componentNode) getLabel() string {
	return n.component.Name
}

/*
	Service Instance
*/
type serviceInstanceNode struct {
	instance *resolve.ComponentInstance
	service  *lang.Service
}

func (n serviceInstanceNode) getGroup() string {
	return "serviceInstance"
}

func (n serviceInstanceNode) getID() string {
	return n.instance.GetKey()
}

func (n serviceInstanceNode) getLabel() string {
	result := fmt.Sprintf(
		`<b>%s</b>
				context: <i>%s</i>`,
		html.EscapeString(n.service.Name),
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
