package lang

import (
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/util"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"regexp"
)

func makeValidator() *validator.Validate {
	result := validator.New()
	_ = result.RegisterValidation("identifier", validateIdentifier)
	_ = result.RegisterValidation("clustertype", validateClusterType)
	_ = result.RegisterValidation("codetype", validateCodeType)
	_ = result.RegisterValidation("expression", validateExpression)
	_ = result.RegisterValidation("template", validateTemplate)
	return result
}

// validateClusterType implements validator.Func and checks if a given string is a valid cluster type
func validateClusterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == "kubernetes"
}

// validateCodeType implements validator.Func and checks if a given string is a valid code type
func validateCodeType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return util.ContainsString([]string{"helm", "aptomi/code/kubernetes-helm"}, value)
}

// validateIdentifier implements validator.Func and checks if a given string (or a list of strings) is a valid identifier(s)
func validateIdentifier(fl validator.FieldLevel) bool {
	field := fl.Field()
	result := true
	if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
		for i := 0; i < field.Len(); i++ {
			result = result && isIdentifier(field.Index(i).Interface().(string))
		}
	} else {
		result = result && isIdentifier(field.String())
	}
	return result
}

// validateExpression implements validator.Func and checks if a given string (or a list of strings) is valid expression(s)
func validateExpression(fl validator.FieldLevel) bool {
	field := fl.Field()
	result := true
	if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
		for i := 0; i < field.Len(); i++ {
			result = result && isExpression(field.Index(i).Interface().(string))
		}
	} else {
		result = result && isExpression(field.String())
	}
	return result
}

// validateTemplate implements validator.Func and checks if a given string (or a list of strings) is valid template(s)
func validateTemplate(fl validator.FieldLevel) bool {
	field := fl.Field()
	result := true
	if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
		for i := 0; i < field.Len(); i++ {
			result = result && isTemplate(field.Index(i).Interface().(string))
		}
	} else {
		result = result && isTemplate(field.String())
	}
	return result
}

func isExpression(expressionStr string) bool {
	_, err := expression.NewExpression(expressionStr)
	return err == nil
}

func isTemplate(templateStr string) bool {
	_, err := template.NewTemplate(templateStr)
	return err == nil
}

func isIdentifier(id string) bool {
	ok, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_-]{0,63}$", id)
	return ok && err == nil
}
