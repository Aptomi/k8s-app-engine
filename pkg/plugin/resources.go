package plugin

type Resources map[string]*ResourceTable

type ResourceTable struct {
	Headers []string
	Items   []Resource
}

type Resource = []string

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
