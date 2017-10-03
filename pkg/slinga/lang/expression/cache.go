package expression

import "sync"

type Cache struct {
	eCache sync.Map
}

func NewCache() *Cache {
	return &Cache{eCache: sync.Map{}}
}

func (cache *Cache) EvaluateAsBool(expressionStr string, params *Parameters) (bool, error) {
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
