package api

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/go-playground/validator"
	"net/http"
)

type jsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Error []jsonError `json:"errors"`
}

func handleError(w http.ResponseWriter, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		writeError(w, buildValidationErrors(validationErrs), http.StatusBadRequest)
		return
	}

	switch err {
	case files.ErrDuplicateFile:
		writeError(w, buildErrors(err, "DuplicateFileError"), http.StatusBadRequest)
	case files.ErrFileNotRegistered:
		writeError(w, buildErrors(err, "FileNotRegistered"), http.StatusNotFound)
	case files.ErrFileNotInCreatedState:
		writeError(w, buildErrors(err, "FileStateError"), http.StatusConflict)
	case files.ErrFileNotInUploadedState:
		writeError(w, buildErrors(err, "FileStateError"), http.StatusConflict)
	case files.ErrNoFilesInCollection:
		writeError(w, buildErrors(err, "EmptyCollection"), http.StatusNotFound)
	default:
		writeError(w, buildErrors(err, "InternalError"), http.StatusInternalServerError)
	}
}

func buildValidationErrors(validationErrs validator.ValidationErrors) jsonErrors {
	jsonErrs := jsonErrors{Error: []jsonError{}}

	for _, validationErr := range validationErrs {
		desc := fmt.Sprintf("%s %s", validationErr.Field(), validationErr.Tag())
		jsonErrs.Error = append(jsonErrs.Error, jsonError{Code: "ValidationError", Description: desc})
	}
	return jsonErrs
}

func writeError(w http.ResponseWriter, errs jsonErrors, httpCode int) {
	encoder := json.NewEncoder(w)
	w.WriteHeader(httpCode)
	encoder.Encode(&errs)
}

func buildErrors(err error, code string) jsonErrors {
	return jsonErrors{Error: []jsonError{{Description: err.Error(), Code: code}}}
}
