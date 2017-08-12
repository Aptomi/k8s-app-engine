package language

import "github.com/Aptomi/aptomi/pkg/slinga/language/expression"

// Context for a given service
type Context struct {
	*SlingaObject

	Criteria   *Criteria
	Labels     *LabelOperations
	Allocation *Allocation
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(params *expression.ExpressionParameters) bool {
	return context.Criteria == nil || context.Criteria.allows(params)
}

func (context *Context) GetObjectType() SlingaObjectType {
	return TypePolicy
}
