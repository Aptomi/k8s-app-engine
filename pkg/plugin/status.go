package plugin

type DeploymentStatus map[string]*ResourceTable

type ResourceTable struct {
	Headers []string
	Items   []Resource
}

type Resource = []string

func (status DeploymentStatus) Merge(with DeploymentStatus) {
	for key, withTable := range with {
		table, exist := status[key]
		if !exist {
			status[key] = withTable
		} else {
			table.Items = append(table.Items, withTable.Items...)
		}
	}
}
