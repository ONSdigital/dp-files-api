package api

import (
	"github.com/go-playground/validator"
	"regexp"
)

func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z]{1}", fl.Field().String())

	return matched
}
