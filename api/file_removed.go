package api

import (
	"context"
	"net/http"
	"strings"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

type RemoveFile func(ctx context.Context, path string) error

func HandleRemoveFile(removeFile RemoveFile, createFileEvent CreateFileEvent, getFileMetadata GetFileMetadata, authMiddleware auth.Middleware, identityClient *clientsidentity.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		path := mux.Vars(req)["path"]

		logData := log.Data{
			"method": req.Method,
			"path":   path,
		}

		accessToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)

		authEntityData, err := getAuthEntityData(ctx, authMiddleware, identityClient, accessToken, logData)
		if err != nil {
			if strings.Contains(err.Error(), "key id unknown or invalid") || strings.Contains(err.Error(), "jwt token is malformed") {
				writeError(w, buildGenericError("Unauthorised", "the request was not authorised"), http.StatusUnauthorized)
				return
			}
			writeError(w, buildGenericError("Forbidden", "the request was not authorised - check token and user's permissions"), http.StatusForbidden)
			return
		}

		logData["entity_data"] = authEntityData

		identityType := log.USER
		if authEntityData.IsServiceAuth {
			identityType = log.SERVICE
		}
		logAuthOption := log.Auth(identityType, authEntityData.EntityData.UserID)

		fileMetadata, err := getFileMetadata(ctx, path)
		if err != nil {
			log.Error(ctx, "failed to get file metadata", err, logData)
			handleError(w, err)
			return
		}

		fileEvent := &files.FileEvent{
			RequestedBy: &files.RequestedBy{
				ID: authEntityData.EntityData.UserID,
			},
			Action:   files.ActionDelete,
			Resource: path,
			File:     &fileMetadata,
		}

		if err := createFileEvent(ctx, fileEvent); err != nil {
			log.Error(ctx, "failed to create file event", err, log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
			handleError(w, err)
			return
		}
		log.Info(ctx, "successfully created file event for file removal", log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)

		if err := removeFile(ctx, path); err != nil {
			log.Error(ctx, "failed to remove file", err, logData)
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
