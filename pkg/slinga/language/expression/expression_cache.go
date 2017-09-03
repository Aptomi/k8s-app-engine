package expression

import "sync"

type ExpressionCache struct {
	eCache sync.Map
}

func NewExpressionCache() *ExpressionCache {
	return &ExpressionCache{eCache: sync.Map{}}
}

func (cache *ExpressionCache) EvaluateAsBool(expressionStr string, params *ExpressionParameters) (bool, error) {
	// Look up expression from the cache
	var expression *Expression
	expressionCached, ok := cache.eCache.Load(expressionStr)
	if ok {
		expression = expressionCached.(*Expression)
	} else {
		// Compile expression, if not found
		// This might happen a several times in parallel, that's okay
		var err error
		expression, err = NewExpression(expressionStr)
		if err != nil {
			return false, err
		}
		cache.eCache.Store(expressionStr, expression)
	}

	// Evaluate expression
	// This seems to be thread safe
	return expression.EvaluateAsBool(params)
}
