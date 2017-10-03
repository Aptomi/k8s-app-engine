package lang

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/template"
)

// Allocation
type Allocation struct {
	Service string
	Keys    []string
}

// Context
type Context struct {
	Name         string
	Criteria     *Criteria
	ChangeLabels LabelOperations `yaml:"change-labels"`
	Allocation   *Allocation
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(params *expression.Parameters, cache *expression.Cache) (bool, error) {
	if context.Criteria == nil {
		return true, nil
	}
	return context.Criteria.allows(params, cache)
}

// Resolves dynamic allocation keys
func (context *Context) ResolveKeys(params *template.Parameters, cache *template.Cache) ([]string, error) {
	if cache == nil {
		cache = template.NewCache()
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
