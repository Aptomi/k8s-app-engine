package expression

type ExpressionCache map[string]*Expression

func NewExpressionCache() ExpressionCache {
	return make(map[string]*Expression)
}

func (cache ExpressionCache) EvaluateAsBool(expressionStr string, params *ExpressionParameters) (bool, error) {
	// Look up expression from cache or compile
	var expr *Expression
	var ok bool
	expr, ok = cache[expressionStr]
	if !ok {
		var err error
		expr, err = NewExpression(expressionStr)
		if err != nil {
			return false, err
		}
		cache[expressionStr] = expr
	}

	// Evaluate
	return expr.EvaluateAsBool(params)
}
