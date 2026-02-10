package api

import (
	"context"
	"encoding/json"
	"net/http"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/gorilla/mux"
)

type GetFileMetadata func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error)

func HandleGetFileMetadata(getMetadata GetFileMetadata) http.HandlerFunc {
	return HandleGetFileMetadataWithAuth(getMetadata, nil)
}

func HandleGetFileMetadataWithAuth(getMetadata GetFileMetadata, authMiddleware auth.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")
		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}

		writeResponse := func(w http.ResponseWriter) {
			if err := json.NewEncoder(w).Encode(metadata); err != nil {
				handleError(w, err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}

		if authMiddleware == nil {
			writeResponse(w)
			return
		}

		attributes := datasetEditionAttributes(metadata)
		authMiddleware.RequireWithAttributes(
			"static-files:read",
			func(w http.ResponseWriter, _ *http.Request) { writeResponse(w) },
			func(_ *http.Request) (map[string]string, error) { return attributes, nil },
		)(w, req)
	}
}

func datasetEditionAttributes(metadata files.StoredRegisteredMetaData) map[string]string {
	attributes := map[string]string{}
	if metadata.ContentItem == nil {
		return attributes
	}
	if metadata.ContentItem.DatasetID == "" || metadata.ContentItem.Edition == "" {
		return attributes
	}
	attributes["dataset_edition"] = metadata.ContentItem.DatasetID + "/" + metadata.ContentItem.Edition
	return attributes
}
