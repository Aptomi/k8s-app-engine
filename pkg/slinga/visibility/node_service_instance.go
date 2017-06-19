package visibility

import "fmt"

type serviceInstanceNode struct {
	serviceName string
	context     string
	allocation  string
}

func newServiceInstanceNode(serviceName string, context string, allocation string) graphNode {
	return serviceInstanceNode{serviceName: serviceName, context: context, allocation: allocation}
}

func (n serviceInstanceNode) getID() string {
	return fmt.Sprintf("svc-inst-%s-%s-%s", n.serviceName, n.context, n.allocation)
}

func (n serviceInstanceNode) getLabel() string {
	return fmt.Sprintf(
		`components: %s
				cluster: %s
				time: %s`,
		"TBD", "TBD", "TBD")
}

func (n serviceInstanceNode) getGroup() string {
	return "serviceInstance"
}

func (n serviceInstanceNode) getEdgeLabel(dst graphNode) string {
	return ""
}
