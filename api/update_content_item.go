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

type UpdateContentItem func(ctx context.Context, path string, contentItem *files.StoredContentItem) error

type ContentItemChange struct {
	ContentItem *ContentItem `json:"content_item"`
}

func HandlerUpdateContentItem(updateContentItem UpdateContentItem, createFileEvent CreateFileEvent, getFileMetadata GetFileMetadata, authMiddleware auth.Middleware, identityClient *clientsidentity.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		path := mux.Vars(req)["path"]

		logData := log.Data{
			"method": req.Method,
			"path":   path,
		}

		var contentItemChange = ContentItemChange{}
		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&contentItemChange); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
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
			Action:   files.ActionUpdate,
			Resource: path,
			File:     &fileMetadata,
		}

		if err := createFileEvent(ctx, fileEvent); err != nil {
			log.Error(ctx, "failed to create file event", err, log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
			handleError(w, err)
			return
		}
		log.Info(ctx, "successfully created file event for content item update", log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)

		updatedContentItem := &files.StoredContentItem{
			DatasetID: contentItemChange.ContentItem.DatasetID,
			Edition:   contentItemChange.ContentItem.Edition,
			Version:   contentItemChange.ContentItem.Version,
		}

		if err := updateContentItem(ctx, path, updatedContentItem); err != nil {
			handleError(w, err)
			return
		}

		fileMetadata.ContentItem = updatedContentItem

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fileMetadata); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
