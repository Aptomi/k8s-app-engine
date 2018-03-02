package plugin

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
