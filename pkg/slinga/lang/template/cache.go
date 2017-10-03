package template

import "sync"

type Cache struct {
	tCache sync.Map
}

func NewCache() *Cache {
	return &Cache{tCache: sync.Map{}}
}

func (cache *Cache) Evaluate(templateStr string, params *Parameters) (string, error) {
	// Look up template from the cache
	var template *Template
	templateCached, ok := cache.tCache.Load(templateStr)
	if ok {
		template = templateCached.(*Template)
	} else {
		// Compile template, if not found
		// This might happen a several times in parallel, that's okay
		var err error
		template, err = NewTemplate(templateStr)
		if err != nil {
			return "", err
		}
		cache.tCache.Store(templateStr, template)
	}

	// Evaluate template
	// This is thread safe. Multiple executions of the same template can execute safely in parallel
	return template.Evaluate(params)
}
