package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-files-api/store"

	"github.com/go-playground/validator"
)

type jsonError struct {
	Code        string `json:"errorCode"`
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
		writeError(w, buildErrors(err, "DuplicateFileError"), http.StatusConflict)
	case store.ErrCollectionIDAlreadySet:
		writeError(w, buildErrors(err, "CollectionIDAlreadySet"), http.StatusBadRequest)
	case store.ErrBundleIDAlreadySet:
		writeError(w, buildErrors(err, "BundleIDAlreadySet"), http.StatusBadRequest)
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
	case store.ErrBothCollectionAndBundleIDSet:
		writeError(w, buildErrors(err, "BothCollectionAndBundleIDSet"), http.StatusBadRequest)
	case store.ErrFileMoved:
		writeError(w, buildErrors(err, "FileMoved"), http.StatusConflict)
	case store.ErrFileIsPublished:
		writeError(w, buildErrors(err, "FileIsPublished"), http.StatusConflict)
	case store.ErrPathNotFound:
		writeError(w, buildErrors(err, "NotFound"), http.StatusNotFound)
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
