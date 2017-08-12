package template

type TemplateCache map[string]*Template

func NewTemplateCache() TemplateCache {
	return make(map[string]*Template)
}

func (cache TemplateCache) Evaluate(templateStr string, params *TemplateParameters) (string, error) {
	// Look up Template from cache or compile
	var templ *Template
	var ok bool
	templ, ok = cache[templateStr]
	if !ok {
		var err error
		templ, err = NewTemplate(templateStr)
		if err != nil {
			return "", err
		}
		cache[templateStr] = templ
	}

	// Evaluate
	return templ.Evaluate(params)
}
