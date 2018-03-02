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

type ResourceTypeHandler func(obj interface{}) []string

type ResourceRegistry struct {
	headers  map[string][]string
	handlers map[string]ResourceTypeHandler
}

func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		make(map[string][]string),
		make(map[string]ResourceTypeHandler),
	}
}

func (reg *ResourceRegistry) AddHandler(resourceType string, headers []string, handler ResourceTypeHandler) {
	if _, exist := reg.headers[resourceType]; exist {
		panic(fmt.Sprintf("duplicate resource type registered: %s", resourceType))
	}

	reg.headers[resourceType] = headers
	reg.handlers[resourceType] = handler
}

func (reg *ResourceRegistry) IsSupported(resourceType string) bool {
	_, ok := reg.headers[resourceType]
	return ok
}

func (reg *ResourceRegistry) Headers(resourceType string) []string {
	return reg.headers[resourceType]
}

func (reg *ResourceRegistry) Handle(resourceType string, obj interface{}) []string {
	return reg.handlers[resourceType](obj)
}
