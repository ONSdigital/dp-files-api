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

type MarkMovementComplete func(ctx context.Context, change files.FileEtagChange) error

func HandleMarkFileMoved(markMovementComplete MarkMovementComplete, createFileEvent CreateFileEvent, getFileMetadata GetFileMetadata, authMiddleware auth.Middleware, idClient *clientsidentity.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		path := mux.Vars(req)["path"]

		logData := log.Data{
			"method": req.Method,
			"path":   path,
		}

		accessToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)
		if accessToken == "" {
			log.Info(ctx, "authorisation failed: no authorisation header in request", log.Classification(log.ProtectiveMonitoring), logData)
			writeError(w, buildGenericError("Unauthorised", "The user is unauthorised"), http.StatusUnauthorized)
			return
		}

		authEntityData, err := getAuthEntityData(ctx, authMiddleware, idClient, accessToken, logData)
		if err != nil {
			log.Error(ctx, "failed to get auth entity data", err, logData)
			if strings.Contains(err.Error(), "key id unknown or invalid") || strings.Contains(err.Error(), "jwt token is malformed") {
				writeError(w, buildGenericError("Unauthorised", "the request was not authorised"), http.StatusUnauthorized)
				return
			}
			writeError(w, buildGenericError("Forbidden", "the request was not authorised - check token and user's permissions"), http.StatusForbidden)
			return
		}

		m := EtagChange{}
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		fileMetadata, err := getFileMetadata(ctx, path)
		if err != nil {
			log.Error(ctx, "failed to get file metadata for audit record", err, logData)
			handleError(w, err)
			return
		}

		if err := createAuditEvent(ctx, createFileEvent, authEntityData.EntityData, authEntityData.IsServiceAuth, files.ActionUpdate, path, &fileMetadata, logData); err != nil {
			handleError(w, err)
			return
		}

		if err := markMovementComplete(ctx, generateFileEtagChange(m, path)); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
