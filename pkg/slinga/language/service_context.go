package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
)

// Context for a given service
type Context struct {
	*SlingaObject

	Criteria     *Criteria
	ChangeLabels LabelOperations `yaml:"change-labels"`
	Allocation *struct {
		Name string
		Keys []string
	}
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(params *expression.ExpressionParameters, cache expression.ExpressionCache) bool {
	return context.Criteria == nil || context.Criteria.allows(params, cache)
}

func (context *Context) GetObjectType() SlingaObjectType {
	return TypePolicy
}
