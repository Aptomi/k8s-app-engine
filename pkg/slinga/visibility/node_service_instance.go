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
}

func newServiceInstanceNode(service *slinga.Service, context string, allocation string, instance *slinga.ComponentInstance) graphNode {
	return serviceInstanceNode{service: service, context: context, allocation: allocation, instance: instance}
}

func (n serviceInstanceNode) getID() string {
	return fmt.Sprintf("svc-inst-%s-%s-%s", n.service.Name, n.context, n.allocation)
}

func (n serviceInstanceNode) getLabel() string {
	return fmt.Sprintf(
		`components: %d
				cluster: %s
				running: %s`,
		len(n.service.Components), // TODO: fix
		n.instance.CalculatedLabels.Labels["cluster"],
		time.NewDiff(n.instance.GetRunningTime()).Humanize(),
	)
}

func (n serviceInstanceNode) getGroup() string {
	return "serviceInstance"
}

func (n serviceInstanceNode) getEdgeLabel(dst graphNode) string {
	return ""
}
