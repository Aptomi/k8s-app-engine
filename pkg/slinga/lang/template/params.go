package template

// Parameters is a set of named parameters for the text template
type Parameters struct {
	params interface{}
}

// NewParams creates a new instance of Parameters
func NewParams(params interface{}) *Parameters {
	return &Parameters{params: params}
}
