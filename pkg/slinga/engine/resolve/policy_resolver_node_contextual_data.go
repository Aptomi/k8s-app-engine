package resolve

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/template"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

/*
	Data exposed to context expressions/criterias defined in policy
*/

// This method defines which contextual information will be exposed to the expression engine (for evaluating criterias)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForContextExpression() *expression.ExpressionParameters {
	return expression.NewExpressionParams(
		node.labels.Labels,
		map[string]interface{}{},
	)
}

/*
	Data exposed to rules defined
*/

// This method defines which contextual information will be exposed to the expression engine (for evaluating rules)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForRuleExpression() *expression.ExpressionParameters {
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
func (node *resolutionNode) getContextualDataForContextAllocationTemplate() *template.TemplateParameters {
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
			Labels:    node.labels.Labels,
			Discovery: node.proxyDiscovery(node.discoveryTreeNode, node.componentKey),
		},
	)
}

/*
	Proxy functions
*/

// How service is visible from the policy language
func (node *resolutionNode) proxyService(service *lang.Service) interface{} {
	return struct {
		lang.Metadata
		Labels interface{}
		Owner  interface{}
	}{
		Metadata: service.Metadata,
		Labels:   service.Labels,
		Owner:    node.proxyUser(node.resolver.externalData.UserLoader.LoadUserByID(service.Owner)),
	}
}

// How user is visible from the policy language
func (node *resolutionNode) proxyUser(user *lang.User) interface{} {
	return struct {
		ID      interface{}
		Name    interface{}
		Labels  interface{}
		Secrets interface{}
	}{
		ID:      user.ID,
		Name:    user.Name,
		Labels:  user.Labels,
		Secrets: node.resolver.externalData.SecretLoader.LoadSecretsByUserID(user.ID),
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
