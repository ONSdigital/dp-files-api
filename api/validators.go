package api

import (
	"regexp"

	"github.com/go-playground/validator"
)

func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString("^[0-9a-zA-Z]", fl.Field().String())

	return matched
}
