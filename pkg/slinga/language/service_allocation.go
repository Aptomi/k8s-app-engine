package language

import "github.com/Aptomi/aptomi/pkg/slinga/language/template"

// Allocation defines within a Context for a given service
type Allocation struct {
	Name         string
	ChangeLabels *LabelOperations `yaml:"change-labels"`
}

// ResolveName resolves name for an allocation
func (allocation *Allocation) ResolveName(parameters *template.TemplateParameters, cache template.TemplateCache) (string, error) {
	if cache == nil {
		cache = template.NewTemplateCache()
	}
	return cache.Evaluate(allocation.Name, parameters)
}
