package engine

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
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
