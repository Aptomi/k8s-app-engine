package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ContextObject = &ObjectInfo{
	Kind:        Kind("context"),
	Constructor: func() BaseObject { return &Context{} },
}

// Context for a given service
type Context struct {
	Metadata

	Criteria     *Criteria
	ChangeLabels LabelOperations `yaml:"change-labels"`
	Allocation   *struct {
		Keys []string
	}
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(params *expression.ExpressionParameters, cache expression.ExpressionCache) (bool, error) {
	if context.Criteria == nil {
		return true, nil
	}
	return context.Criteria.allows(params, cache)
}

// Resolves dynamic allocation keys
func (context *Context) ResolveKeys(params *template.TemplateParameters, cache template.TemplateCache) ([]string, error) {
	if cache == nil {
		cache = template.NewTemplateCache()
	}
	// Resolve allocation keys (they can be dynamic, depending on user labels)
	result := []string{}
	for _, key := range context.Allocation.Keys {
		keyResolved, err := cache.Evaluate(key, params)
		if err != nil {
			return nil, err
		}
		result = append(result, keyResolved)
	}
	return result, nil
}
