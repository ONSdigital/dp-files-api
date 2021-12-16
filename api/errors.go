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
	if verrs, ok := err.(validator.ValidationErrors); ok {
		jerrs := jsonErrors{
			Error: []jsonError{},
		}
		for _, verr := range verrs {
			desc := fmt.Sprintf("%s %s", verr.Field(), verr.Tag())
			jerrs.Error = append(jerrs.Error, jsonError{Code: "ValidationError", Description: desc})
		}

		encoder := json.NewEncoder(w)

		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(&jerrs)

		return
	}

	switch err {
	case files.ErrDuplicateFile:
		writeError(w, err, "DuplicateFileError", http.StatusBadRequest)
	case files.ErrFileNotRegistered:
		writeError(w, err, "FileNotRegistered", http.StatusNotFound)
	case files.ErrFileNotInCreatedState:
		writeError(w, err, "FileStateError", http.StatusConflict)
	default:
		writeError(w, err, "DatabaseError", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, err error, code string, httpCode int) {
	encoder := json.NewEncoder(w)

	errs := jsonErrors{Error: []jsonError{{Description: err.Error(), Code: code}}}
	w.WriteHeader(httpCode)
	encoder.Encode(&errs)
}
