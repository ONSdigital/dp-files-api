package api

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-playground/validator"
)

func mimeValidator(fl validator.FieldLevel) bool {
	mt := fl.Field().String()
	mtype := mimetype.Lookup(mt)

	return mtype != nil
}
