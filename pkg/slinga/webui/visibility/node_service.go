package visibility

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
)

type serviceNode struct {
	serviceName string
}

func newServiceNode(serviceName string) graphNode {
	return serviceNode{serviceName: serviceName}
}

func (n serviceNode) getIDPrefix() string {
	return "svc-"
}

func (n serviceNode) getGroup() string {
	return "service"
}

func (n serviceNode) getID() string {
	return fmt.Sprintf("%s%s", n.getIDPrefix(), n.serviceName)
}

func (n serviceNode) isItMyID(id string) string {
	return cutPrefixOrEmpty(id, n.getIDPrefix())
}

func (n serviceNode) getLabel() string {
	return n.serviceName
}

func (n serviceNode) getEdgeLabel(dst graphNode) string {
	// if it's an edge from service to service instance, write context information on it
	if dstInst, ok := dst.(serviceInstanceNode); ok {
		return dstInst.contextWithKeys
	}
	return ""
}

func (n serviceNode) getDetails(id string, state *resolve.ResolvedState) interface{} {
	return state.Policy.Services[id]
}
