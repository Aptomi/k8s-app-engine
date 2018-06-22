package plugin

import "fmt"

// Resources represents description of all deployed on a cluster resources of different types
type Resources map[string]*ResourceTable

// ResourceTable is a list of resources of the same type as columns with column headers
type ResourceTable struct {
	Headers []string
	Items   []Resource
}

// Resource is a list of columns representing deployed resources
type Resource = []string

// Merge takes Resources and merges them into current resources object
func (status Resources) Merge(with Resources) {
	for key, withTable := range with {
		table, exist := status[key]
		if !exist {
			status[key] = withTable
		} else {
			table.Items = append(table.Items, withTable.Items...)
		}
	}
}

// ResourceTypeHandler represents function that converts object into columns
type ResourceTypeHandler func(obj interface{}) []string

// ResourceRegistry helps to store and use handlers and headers for resources
type ResourceRegistry struct {
	headers  map[string][]string
	handlers map[string]ResourceTypeHandler
}

// NewResourceRegistry creates new ResourceRegistry
func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		headers:  make(map[string][]string),
		handlers: make(map[string]ResourceTypeHandler),
	}
}

// AddHandler adds specified resource type handler to registry by specified resource type with specified headers
func (reg *ResourceRegistry) AddHandler(resourceType string, headers []string, handler ResourceTypeHandler) {
	if _, exist := reg.headers[resourceType]; exist {
		panic(fmt.Sprintf("duplicate resource type registered: %s", resourceType))
	}

	reg.headers[resourceType] = headers
	reg.handlers[resourceType] = handler
}

// IsSupported checks if specified resource type supported by registry
func (reg *ResourceRegistry) IsSupported(resourceType string) bool {
	_, ok := reg.headers[resourceType]
	return ok
}

// Headers returns headers for specified resource type
func (reg *ResourceRegistry) Headers(resourceType string) []string {
	return reg.headers[resourceType]
}

// Handle returns columns for specified object with specified resource type
func (reg *ResourceRegistry) Handle(resourceType string, obj interface{}) []string {
	return reg.handlers[resourceType](obj)
}
