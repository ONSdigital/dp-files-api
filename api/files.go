package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-files-api/files"
)

func CreateFileUploadStartedHandler(creatorFunc files.CreateUploadStartedEntry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		m := files.MetaData{}

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			return
		}

		m.CreatedAt = time.Now()
		m.LastModified = time.Now()
		m.State = "CREATED"

		encoder := json.NewEncoder(w)

		type jsonError struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		}

		type JsonErrors struct {
			Error []jsonError `json:"errors"`
		}

		err = creatorFunc(req.Context(), m)
		if err != nil {
			if err == files.ErrDuplicateFile {
				errs := JsonErrors{Error: []jsonError{{Description: err.Error(), Code: "DuplicateFileError"}}}
				w.WriteHeader(http.StatusBadRequest)
				encoder.Encode(&errs)
				return
			}
			errs := JsonErrors{Error: []jsonError{{Description: err.Error(), Code: "DatabaseError"}}}
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(&errs)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
