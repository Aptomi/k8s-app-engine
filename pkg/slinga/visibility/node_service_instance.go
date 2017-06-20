package visibility

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga"
	"github.com/Frostman/aptomi/pkg/slinga/time"
)

type serviceInstanceNode struct {
	service    *slinga.Service
	context    string
	allocation string
	instance   *slinga.ComponentInstance
	primary    bool
}

func newServiceInstanceNode(service *slinga.Service, context string, allocation string, instance *slinga.ComponentInstance, primary bool) graphNode {
	return serviceInstanceNode{
		service:    service,
		context:    context,
		allocation: allocation,
		instance:   instance,
		primary:    primary,
	}
}

func (n serviceInstanceNode) getID() string {
	return fmt.Sprintf("svc-inst-%s-%s-%s", n.service.Name, n.context, n.allocation)
}

func (n serviceInstanceNode) getLabel() string {
	if n.primary {
		return fmt.Sprintf(
			`%s
				components: %d
				cluster: %s
				running: %s`,
			n.service.Name,
			len(n.service.Components), // TODO: fix
			n.instance.CalculatedLabels.Labels["cluster"],
			time.NewDiff(n.instance.GetRunningTime()).Humanize(),
		)
	}
	return fmt.Sprintf(
		`%s
			cluster: %s
			running: %s`,
		n.service.Name,
		n.instance.CalculatedLabels.Labels["cluster"],
		time.NewDiff(n.instance.GetRunningTime()).Humanize(),
	)
}

func (n serviceInstanceNode) getGroup() string {
	if n.primary {
		return "serviceInstancePrimary"
	}
	return "serviceInstance"
}

func (n serviceInstanceNode) getEdgeLabel(dst graphNode) string {
	return ""
}
