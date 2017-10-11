package config

import (
	"github.com/asaskevich/govalidator"
)

func Validate(config Base) (bool, error) {
	return govalidator.ValidateStruct(config)
}
