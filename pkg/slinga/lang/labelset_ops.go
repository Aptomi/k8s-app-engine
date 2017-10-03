package lang

// LabelOperations defines the set of label manipulations (e.g. set/remove)
type LabelOperations map[string]map[string]string

// NewLabelOperations creates a new LabelOperations object, given "set" and "remove" parameters
func NewLabelOperations(setMap map[string]string, removeMap map[string]string) LabelOperations {
	result := LabelOperations{}
	result["set"] = setMap
	result["remove"] = removeMap
	return result
}

// NewLabelOperations creates a new LabelOperations object, to set a single "k"="v" label
func NewLabelOperationsSetSingleLabel(k string, v string) LabelOperations {
	result := LabelOperations{}
	result["set"] = map[string]string{k: v}
	return result
}
