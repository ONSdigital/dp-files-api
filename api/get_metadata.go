package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

type GetFileMetadata func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error)

func HandleGetFileMetadataWithAuth(getMetadata GetFileMetadata, authMiddleware auth.Middleware, idClient *clientsidentity.Client, permissionsChecker auth.PermissionsChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")

		logData := log.Data{
			"method": req.Method,
			"path":   req.URL.Path,
		}

		accessToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)
		if accessToken == "" {
			log.Info(req.Context(), "authorisation failed: no authorisation header in request", log.Classification(log.ProtectiveMonitoring), logData)
			writeError(w, buildGenericError("Unauthorised", "The user is unauthorised"), http.StatusUnauthorized)
			return
		}

		authEntityData, err := getAuthEntityData(req.Context(), authMiddleware, idClient, accessToken, logData)
		if err != nil {
			log.Error(req.Context(), "the request was not authorised", err, logData)
			if strings.Contains(err.Error(), "key id unknown or invalid") || strings.Contains(err.Error(), "jwt token is malformed") {
				writeError(w, buildGenericError("Unauthorised", "the request was not authorised"), http.StatusUnauthorized)
				return
			}
			writeError(w, buildGenericError("Forbidden", "the request was not authorised - check token and user's permissions"), http.StatusForbidden)
			return
		}

		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			log.Error(req.Context(), "unable to retrieve metadata", err)
			handleError(w, err)
			return
		}
		var permissionAttrs map[string]string
		if metadata.ContentItem != nil {
			permissionAttrs = map[string]string{
				"dataset_edition": metadata.ContentItem.DatasetID + "/" + metadata.ContentItem.Edition,
			}
		}

		logData = log.Data{
			"entity_data": authEntityData,
		}

		if checkUserPermission(req, logData, "static-files:read", permissionAttrs, permissionsChecker, authEntityData.EntityData) {
			if err := json.NewEncoder(w).Encode(metadata); err != nil {
				handleError(w, err)
				return
			}
			w.WriteHeader(http.StatusOK)
			log.Info(req.Context(), "GetFileMetadata endpoint: get file metadata request successful", logData)
		} else {
			logData["message"] = "user/service does not have required permission"

			if authEntityData.IsServiceAuth {
				log.Info(req.Context(), "authorisation failed: request has no permission", log.Classification(log.ProtectiveMonitoring), log.Auth(log.SERVICE, authEntityData.EntityData.UserID), logData)
			} else {
				log.Info(req.Context(), "authorisation failed: request has no permission", log.Classification(log.ProtectiveMonitoring), log.Auth(log.USER, authEntityData.EntityData.UserID), logData)
			}
			writeError(w, buildGenericError("Forbidden", "the request was not authorised - check token and user's permissions"), http.StatusForbidden)
			return
		}
	}
}

func HandleGetFileMetadata(getMetadata GetFileMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")
		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}

		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
