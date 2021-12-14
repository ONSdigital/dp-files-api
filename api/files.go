package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-files-api/files"
	"net/http"
)

type jsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Error []jsonError `json:"errors"`
}

func CreateFileUploadStartedHandler(creatorFunc files.CreateUploadStartedEntry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		m := files.MetaData{}

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			return
		}

		err = creatorFunc(req.Context(), m)
		if err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch err {
	case files.ErrDuplicateFile:
		writeError(w, err, "DuplicateFileError", http.StatusBadRequest)
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
