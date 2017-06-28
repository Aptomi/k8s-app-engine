package language

import "github.com/Frostman/aptomi/pkg/slinga/language/yaml"

// Context for a given service
type Context struct {
	Name        string
	Service     string
	Criteria    *Criteria
	Labels      *LabelOperations
	Allocations []*Allocation
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(labels LabelSet) bool {
	return context.Criteria == nil || context.Criteria.allows(labels)
}

// Loads context from file
func loadContextFromFile(fileName string) *Context {
	return yaml.LoadObjectFromFile(fileName, new(Context)).(*Context)
}
