package lang

import (
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"regexp"
)

func makeValidator() *validator.Validate {
	result := validator.New()
	_ = result.RegisterValidation("identifier", ValidateIdentifier)
	_ = result.RegisterValidation("clustertype", ValidateClusterType)
	_ = result.RegisterValidation("expression", ValidateExpression)
	return result
}

// ValidateIdentifier implements validator.Func and checks if a given string is a valid identifier in Aptomi
func ValidateIdentifier(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	ok, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_-]{0,63}$", value)
	return ok && err == nil
}

// ValidateClusterType implements validator.Func and checks if a given string is a valid cluster type
func ValidateClusterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == "kubernetes"
}

// ValidateExpression implements validator.Func and checks if a given string is valid expression
func ValidateExpression(fl validator.FieldLevel) bool {
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

func isExpression(expressionStr string) bool {
	_, err := expression.NewExpression(expressionStr)
	return err == nil
}
