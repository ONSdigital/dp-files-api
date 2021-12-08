package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-files-api/files"
	"net/http"
	"time"
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

		creatorFunc(req.Context(), m)

		w.WriteHeader(http.StatusCreated)
	}
}
