package visibility

import "fmt"

type serviceNode struct {
	serviceName string
}

func newServiceNode(serviceName string) graphNode {
	return serviceNode{serviceName: serviceName}
}

func (n serviceNode) getID() string {
	return fmt.Sprintf("svc-%s", n.serviceName)
}

func (n serviceNode) getLabel() string {
	return n.serviceName
}

func (n serviceNode) getGroup() string {
	return "service"
}

func (n serviceNode) getEdgeLabel(dst graphNode) string {
	// if it's an edge from service to service instance, write context information on it
	if dstInst, ok := dst.(serviceInstanceNode); ok {
		return fmt.Sprintf("%s/%s", dstInst.context, dstInst.allocation)
	}
	return ""
}
