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

type MarkFilePublished func(ctx context.Context, path string) error

func HandleMarkFilePublished(markFilePublished MarkFilePublished, createFileEvent CreateFileEvent, getFileMetadata GetFileMetadata, authMiddleware auth.Middleware, idClient *clientsidentity.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		path := mux.Vars(req)["path"]

		logData := log.Data{
			"method": req.Method,
			"path":   path,
		}

		accessToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)

		authEntityData, err := getAuthEntityData(ctx, authMiddleware, idClient, accessToken, logData)
		if err != nil {
			log.Error(ctx, "failed to get auth entity data for mark file published", err, logData)
		}

		if err := markFilePublished(ctx, path); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)

		fileMetadata, err := getFileMetadata(ctx, path)
		if err != nil {
			log.Error(ctx, "failed to get file metadata for audit record", err, logData)
			return
		}

		identityType := log.USER
		if authEntityData != nil && authEntityData.IsServiceAuth {
			identityType = log.SERVICE
		}

		var userID string
		if authEntityData != nil && authEntityData.EntityData != nil {
			userID = authEntityData.EntityData.UserID
		}

		logAuthOption := log.Auth(identityType, userID)

		auditEvent := &files.FileEvent{
			RequestedBy: &files.RequestedBy{
				ID: userID,
			},
			Action:   files.ActionUpdate,
			Resource: path,
			File:     &fileMetadata,
		}

		if err := createFileEvent(ctx, auditEvent); err != nil {
			log.Error(ctx, "failed to create audit record for mark file published", err, log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
			return
		}

		log.Info(ctx, "successfully created audit record for mark file published", log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
	}
}
