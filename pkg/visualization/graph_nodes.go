package visualization

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
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
	serviceKey := n.b.resolution.GetDependencyInstanceMap()[runtime.KeyForStorable(n.dependency)]
	if len(serviceKey) > 0 {
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
	timeRunning := ""
	if !n.instance.CreatedAt.IsZero() {
		timeRunning = fmt.Sprintf("\nrunning: <i>%s</i>", html.EscapeString(util.NewTimeDiff(n.instance.GetRunningTime()).Humanize()))
	}
	return fmt.Sprintf(
		`<b>%s</b>
			context: <i>%s</i>
			cluster: <i>%s</i>%s`,
		html.EscapeString(n.service.Name),
		html.EscapeString(n.instance.Metadata.Key.ContextNameWithKeys),
		html.EscapeString(n.instance.CalculatedLabels.Labels[lang.LabelCluster]),
		timeRunning,
	)
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
