package language

import "github.com/Aptomi/aptomi/pkg/slinga/language/template"

// Allocation defines how service is allocated
type Allocation struct {
	Name         string
}

// ResolveName resolves name for an allocation
func (allocation *Allocation) ResolveName(parameters *template.TemplateParameters, cache template.TemplateCache) (string, error) {
	if cache == nil {
		cache = template.NewTemplateCache()
	}
	return cache.Evaluate(allocation.Name, parameters)
}
