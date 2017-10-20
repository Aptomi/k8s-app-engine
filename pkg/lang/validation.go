package lang

import (
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

func makeValidator() *validator.Validate {
	result := validator.New()
	result.RegisterValidation("identifier", ValidateIdentifier)
	result.RegisterValidation("clustertype", ValidateClusterType)
	return result
}

// ValidateIdentifier implements validator.Func and checks if a given string identifier is valid
func ValidateIdentifier(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	ok, err := regexp.MatchString("[a-zA-Z][a-zA-Z0-9]{0,63}", value)
	return ok && err == nil
}

// ValidateClusterType implements validator.Func and checks if a given string is a valid cluster type
func ValidateClusterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == "kubernetes"
}
