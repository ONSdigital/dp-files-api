package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

type CreateFileEvent func(ctx context.Context, event *files.FileEvent) error

func HandlerCreateFileEvent(createFileEvent CreateFileEvent, authMiddleware auth.Middleware, idClient *clientsidentity.Client, permissionsChecker auth.PermissionsChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var event files.FileEvent

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
			if strings.Contains(err.Error(), "key id unknown or invalid") || strings.Contains(err.Error(), "jwt token is malformed") {
				writeError(w, buildGenericError("Unauthorised", "the request was not authorised"), http.StatusUnauthorized)
				return
			}
			writeError(w, buildGenericError("Forbidden", "the request was not authorised - check token and user's permissions"), http.StatusForbidden)
			return
		}

		if err := json.NewDecoder(req.Body).Decode(&event); err != nil {
			log.Error(req.Context(), "Failed to decode JSON request body", err, logData)
			writeError(w, buildGenericError("BadJson", "The JSON is not in a valid format"), http.StatusBadRequest)
			return
		}

		if err := validateFileEvent(&event); err != nil {
			log.Error(req.Context(), "File event validation failed", err, logData)
			writeError(w, buildGenericError("InvalidRequest", "Unable to process request due to a malformed or invalid request body or query parameter"), http.StatusBadRequest)
			return
		}

		var permissionAttrs map[string]string
		if event.File.ContentItem != nil {
			permissionAttrs = map[string]string{
				"dataset_edition": event.File.ContentItem.DatasetID + "/" + event.File.ContentItem.Edition,
			}
		}

		logData["entity_data"] = authEntityData

		if checkUserPermission(req, logData, "static-files:read", permissionAttrs, permissionsChecker, authEntityData.EntityData) {

			if err := createFileEvent(req.Context(), &event); err != nil {
				handleError(w, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			if err := json.NewEncoder(w).Encode(&event); err != nil {
				handleError(w, err)
				return
			}

			log.Info(req.Context(), "CreateFileEvent endpoint: create file event request successful", logData)
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

func validateFileEvent(event *files.FileEvent) error {
	if event.RequestedBy == nil || event.RequestedBy.ID == "" {
		return errors.New("requested_by.id is required")
	}
	if event.Action == "" {
		return errors.New("action is required")
	}
	if event.Resource == "" {
		return errors.New("resource is required")
	}
	if event.File == nil || event.File.Path == "" {
		return errors.New("file.path is required")
	}
	return nil
}

func buildGenericError(code, description string) JSONErrors {
	return JSONErrors{Error: []JSONError{{Code: code, Description: description}}}
}
