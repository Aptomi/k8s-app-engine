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
func (node *resolutionNode) proxyDiscovery(discoveryTree NestedParameterMap, componentKey string) interface{} {
	result := discoveryTree.MakeCopy()
	result["instance"] = EscapeName(componentKey)
	return result
}
