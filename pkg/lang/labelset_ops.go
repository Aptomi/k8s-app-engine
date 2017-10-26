package lang

// LabelOperations defines label transform operations. It supports two types of operations - 'set' and 'remove',
// those strings are used as keys in the main map.
//
// When 'set' is used, the inner name->value map defines which labels should be added/overwritten.
// When 'remove' is used, the inner name->value map defines which labels should be deleted. In this case value
// doesn't matter and name can even point to an empty string as value
//
// The typical usage of this struct is to take LabelSet and transform it using LabelOperations.
type LabelOperations map[string]map[string]string

// NewLabelOperations creates a new LabelOperations object, given "set" and "remove" parameters
func NewLabelOperations(setMap map[string]string, removeMap map[string]string) LabelOperations {
	result := LabelOperations{}
	result["set"] = setMap
	result["remove"] = removeMap
	return result
}

// NewLabelOperationsSetSingleLabel creates a new LabelOperations object to set a single "k"="v" label
func NewLabelOperationsSetSingleLabel(k string, v string) LabelOperations {
	result := LabelOperations{}
	result["set"] = map[string]string{k: v}
	return result
}
