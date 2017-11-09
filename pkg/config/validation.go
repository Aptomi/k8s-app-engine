package config

import (
	"fmt"
	english "github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/go-playground/validator.v9/translations/en"
	"os"
	"reflect"
	"strings"
)

// Custom error for config validation
type configValidationError struct {
	errList []string
}

func (err configValidationError) Error() string {
	return strings.Join(err.errList, "\n")
}

func (err *configValidationError) addError(errStr string) {
	err.errList = append(err.errList, errStr)
}

// Validator is a custom validator for configs
type Validator struct {
	val    *validator.Validate
	config Base
	trans  ut.Translator
}

// NewValidator creates a new Validator
func NewValidator(config Base) *Validator {
	result := validator.New()

	// independent validators
	_ = result.RegisterValidation("dir", validateDir)
	_ = result.RegisterValidation("file", validateFile)

	// default translations
	eng := english.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")
	err := en.RegisterDefaultTranslations(result, trans)
	if err != nil {
		panic(err)
	}

	// additional translations
	translations := []struct {
		tag         string
		translation string
	}{
		{
			tag:         "dir",
			translation: fmt.Sprintf("{0} must point to an existing directory, but found '{1}'"),
		},
		{
			tag:         "file",
			translation: fmt.Sprintf("{0} must point to an existing file, but found '{1}'"),
		},
	}
	for _, t := range translations {
		err = result.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation), translateFunc)
		if err != nil {
			panic(err)
		}
	}

	return &Validator{
		val:    result,
		config: config,
		trans:  trans,
	}
}

func registrationFunc(tag string, translation string) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, true); err != nil {
			return
		}
		return
	}
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field(), reflect.ValueOf(fe.Value()).String())
	if err != nil {
		return fe.(error).Error()
	}
	return t
}

// Validate validates config for errors and returns an error (it can be casted to
// configValidationError, containing a list of errors inside). When error is printed as string, it will
// automatically contains the full list of validation errors.
func (v *Validator) Validate() error {
	// validate policy
	err := v.val.Struct(v.config)
	if err == nil {
		return nil
	}

	// collect human-readable errors
	result := configValidationError{}
	vErrors := err.(validator.ValidationErrors)
	for _, vErr := range vErrors {
		errStr := fmt.Sprintf("%s: %s", vErr.Namespace(), vErr.Translate(v.trans))
		result.addError(errStr)
	}

	return result
}

// checks if a given string is an existing directory
func validateDir(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

// checks if a given string is an existing file
func validateFile(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	}
	return false
}
