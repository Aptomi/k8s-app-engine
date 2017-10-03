package template

type TemplateParameters struct {
	params interface{}
}

func NewTemplateParams(params interface{}) *TemplateParameters {
	return &TemplateParameters{params: params}
}
