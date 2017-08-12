package engine

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
)

type EngineCache struct {
	expressionCache expression.ExpressionCache
	templateCache   template.TemplateCache
}

func NewEngineCache() *EngineCache {
	return &EngineCache{
		expressionCache: expression.NewExpressionCache(),
		templateCache: template.NewTemplateCache(),
	}
}
