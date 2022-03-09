package api

import (
	"github.com/go-playground/validator"
	"regexp"
)

func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	matched, _ := regexp.MatchString("^[a-zA-Z]{1}", path)

	return matched
}
