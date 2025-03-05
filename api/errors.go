package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-files-api/store"

	"github.com/go-playground/validator"
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
	case store.ErrDuplicateFile:
		writeError(w, buildErrors(err, "DuplicateFileError"), http.StatusBadRequest)
	case store.ErrCollectionIDAlreadySet:
		writeError(w, buildErrors(err, "CollectionIDAlreadySet"), http.StatusBadRequest)
	case store.ErrFileNotRegistered:
		writeError(w, buildErrors(err, "FileNotRegistered"), http.StatusNotFound)
	case store.ErrFileNotInCreatedState,
		store.ErrFileNotInUploadedState,
		store.ErrCollectionIDNotSet,
		store.ErrFileStateMismatch,
		store.ErrEtagMismatchWhilePublishing:
		writeError(w, buildErrors(err, "FileStateError"), http.StatusConflict)
	case store.ErrNoFilesInCollection:
		writeError(w, buildErrors(err, "EmptyCollection"), http.StatusNotFound)
	case store.ErrFileIsNotPublishable:
		writeError(w, buildErrors(err, "FileNotPublishable"), http.StatusConflict)
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
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(&errs)
}

func buildErrors(err error, code string) jsonErrors {
	return jsonErrors{Error: []jsonError{{Description: err.Error(), Code: code}}}
}
