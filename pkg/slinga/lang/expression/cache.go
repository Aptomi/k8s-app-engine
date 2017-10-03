package expression

import "sync"

// Cache is a cache of compiled expressions
type Cache struct {
	eCache sync.Map
}

// NewCache creates a new Cache
func NewCache() *Cache {
	return &Cache{eCache: sync.Map{}}
}

// EvaluateAsBool evaluates boolean expression given a set of parameters, using the compiled expression from cache
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
