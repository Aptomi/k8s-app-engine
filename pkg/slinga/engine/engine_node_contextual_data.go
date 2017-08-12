package engine

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

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
