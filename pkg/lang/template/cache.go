package template

import "sync"

// Cache is a thread-safe cache of compiled text templates
type Cache struct {
	tCache sync.Map
}

// NewCache creates a new thread-safe Cache
func NewCache() *Cache {
	return &Cache{tCache: sync.Map{}}
}

// Evaluate evaluates text template given a set of parameters.
// If an compiled text template already exists in cache, it will be used.
// Otherwise it will get compiled and added to the cache before evaluating the text template.
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (cache *Cache) Evaluate(templateStr string, params *Parameters) (string, error) {
	// Look up template from the cache
	var template *Template
	templateCached, ok := cache.tCache.Load(templateStr)
	if ok {
		template = templateCached.(*Template) // nolint: errcheck
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
