package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/gorilla/mux"
)

type GetFileMetadata func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error)

func HandleGetFileMetadata(getMetadata GetFileMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")
		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}

		json.NewEncoder(w).Encode(metadata)

		w.WriteHeader(http.StatusOK)
	}
}
