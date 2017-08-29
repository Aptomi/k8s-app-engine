package resolve

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

/*
	Data exposed to expressions defined in policy
*/

// This method defines which contextual information will be exposed to the expression engine (e.g. for evaluating criterias)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForExpression() *expression.ExpressionParameters {
	return expression.NewExpressionParams(
		node.labels.Labels,
		map[string]interface{}{
			"service": node.proxyService(node.service),
		},
	)
}

/*
	Data exposed to templates defined in policy
*/

// This method defines which contextual information will be exposed to the template engine (for evaluating all templates - discovery, code params, etc)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForAllocationTemplate() *template.TemplateParameters {
	return template.NewTemplateParams(
		struct {
			User   interface{}
			Labels interface{}
		}{
			User:   node.proxyUser(node.user),
			Labels: node.labels.Labels,
		},
	)
}

// This method defines which contextual information will be exposed to the template engine (for evaluating all templates - discovery, code params, etc)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForCodeDiscoveryTemplate() *template.TemplateParameters {
	return template.NewTemplateParams(
		struct {
			User      interface{}
			Labels    interface{}
			Discovery interface{}
		}{
			User:      node.proxyUser(node.user),
			Labels:    node.componentLabels.Labels,
			Discovery: node.proxyDiscovery(node.discoveryTreeNode, node.componentKey),
		},
	)
}

/*
	Proxy functions
*/

// How service is visible from the policy language
func (node *resolutionNode) proxyService(service *language.Service) interface{} {
	return struct {
		object.Metadata
		Owner interface{}
	}{
		Metadata: service.Metadata,
		Owner:    node.proxyUser(node.resolver.userLoader.LoadUserByID(service.Owner)),
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
