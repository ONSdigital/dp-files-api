package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/gorilla/mux"
)

const PermissionStaticFilesRead = "static-files:read"

type GetFileMetadata func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error)

func HandleGetFileMetadata(getMetadata GetFileMetadata, authMiddleware auth.Middleware, permission string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		if authMiddleware != nil && permission != "" {
			bearerToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)
			if bearerToken == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if _, err := authMiddleware.Parse(bearerToken); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}

		writeResponse := func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(metadata); err != nil {
				handleError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
		}

		if authMiddleware != nil && permission != "" {
			attributes := permissionAttributesFromMetadata(metadata)
			authHandler := authMiddleware.RequireWithAttributes(permission, writeResponse, func(_ *http.Request) (map[string]string, error) {
				return attributes, nil
			})
			authHandler(w, req)
			return
		}

		writeResponse(w, req)
	}
}

func permissionAttributesFromMetadata(metadata files.StoredRegisteredMetaData) map[string]string {
	attributes := map[string]string{}
	if metadata.ContentItem == nil || metadata.ContentItem.DatasetID == "" {
		return attributes
	}

	if metadata.ContentItem.Edition == "" {
		attributes["dataset_edition"] = metadata.ContentItem.DatasetID
		return attributes
	}

	attributes["dataset_edition"] = metadata.ContentItem.DatasetID + "/" + metadata.ContentItem.Edition
	return attributes
}
