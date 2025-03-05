package api

import (
	"regexp"

	"github.com/go-playground/validator"
)

func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z]{1}", fl.Field().String())

	return matched
}
