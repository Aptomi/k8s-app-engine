package expression

import "sync"

// Cache is a thread-safe cache of compiled expressions
type Cache struct {
	eCache sync.Map
}

// NewCache creates a new thread-safe Cache
func NewCache() *Cache {
	return &Cache{eCache: sync.Map{}}
}

// EvaluateAsBool evaluates boolean expression given a set of parameters.
// If an compiled expression already exists in cache, it will be used.
// Otherwise it will get compiled and added to the cache before evaluating the expression.
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (cache *Cache) EvaluateAsBool(expressionStr string, params *Parameters) (bool, error) {
	// Look up expression from the cache
	var expression *Expression
	expressionCached, ok := cache.eCache.Load(expressionStr)
	if ok {
		expression = expressionCached.(*Expression) // nolint: errcheck
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
