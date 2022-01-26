package api

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-playground/validator"
	"regexp"
)

func mimeValidator(fl validator.FieldLevel) bool {
	mt := fl.Field().String()
	mtype := mimetype.Lookup(mt)

	return mtype != nil
}
func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	matched, _ := regexp.MatchString("^[a-zA-Z]{1}", path)

	return matched
}
