package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

type ServiceViewObject struct {
	serviceName string
	state       slinga.ServiceUsageState
	g           graph
}

func NewServiceViewObject(serviceName string, state slinga.ServiceUsageState) ServiceViewObject {
	return ServiceViewObject{serviceName: serviceName, state: state}
}

func (svo ServiceViewObject) GetData() interface{} {
	// create an empty graph
	g := NewGraph()

	// 1 - add a node with a given service
	svcNode := newServiceNode(svo.serviceName)
	g.addNode(svcNode)

	// 2 - find all instances of a given service. add them as "instance nodes"
	for k, _ := range svo.state.ResolvedUsage.ComponentInstanceMap {
		service, context, allocation, component := slinga.ParseServiceUsageKey(k)
		if service == svo.serviceName && component == slinga.ComponentRootName {
			// add a node with an instance of our service
			svcInstanceNode := newServiceInstanceNode(service, context, allocation)
			g.addNode(svcInstanceNode)

			// connect service node and instance node
			g.addEdge(svcNode, svcInstanceNode)
		}
	}

	// return the resulting graph
	return g.GetData()
}