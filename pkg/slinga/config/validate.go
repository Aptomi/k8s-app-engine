package config

import (
	"github.com/asaskevich/govalidator"
	"os"
)

func init() {
	govalidator.TagMap["dir"] = govalidator.Validator(func(path string) bool {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			return true
		}
		return false
	})
}

func Validate(config Base) (bool, error) {
	return govalidator.ValidateStruct(config)
}
