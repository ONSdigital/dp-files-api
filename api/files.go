package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-files-api/files"
	"net/http"
)

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
