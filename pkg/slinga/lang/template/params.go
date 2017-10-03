package template

type Parameters struct {
	params interface{}
}

func NewParams(params interface{}) *Parameters {
	return &Parameters{params: params}
}
