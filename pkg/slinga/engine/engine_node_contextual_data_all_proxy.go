package engine

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

// How service is visible from the policy language
func (node *resolutionNode) proxyService(service *language.Service) interface{} {
	return struct {
		Metadata interface{}
		Owner    interface{}
	}{
		Metadata: service.Metadata,
		Owner:    node.proxyUser(node.state.userLoader.LoadUserByID(service.Owner)),
	}
}

// How user is visible from the policy language
func (node *resolutionNode) proxyUser(user *language.User) interface{} {
	return struct {
		Name   interface{}
		Labels interface{}
	}{
		Name:   user.Name,
		Labels: user.Labels,
	}
}

// How discovery tree is visible from the policy language
func (node *resolutionNode) proxyDiscovery(discoveryTree NestedParameterMap, cik *ComponentInstanceKey) interface{} {
	result := discoveryTree.MakeCopy()

	// special case to announce own component instance
	result["instance"] = EscapeName(cik.GetKey())

	// special case to announce own component ID
	result["instanceId"] = HashFnv(cik.GetKey())

	// expose parent service information as well
	if cik.IsComponent() {
		// Get service key
		serviceCik := cik.GetParentServiceKey()

		// create a bucket for service
		result["service"] = NestedParameterMap{}

		// special case to announce own component instance
		result.GetNestedMap("service")["instance"] = EscapeName(serviceCik.GetKey())

		// special case to announce own component ID
		result.GetNestedMap("service")["instanceId"] = HashFnv(serviceCik.GetKey())
	}

	return result
}
