package api

import (
	"context"
	"net/http"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
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

		fileMetadata, err := getFileMetadata(ctx, path)
		if err != nil {
			log.Error(ctx, "failed to get file metadata for audit record", err, logData)
			handleError(w, err)
			return
		}

		if statusCode, err := createAuditEvent(ctx, req, authMiddleware, idClient, createFileEvent, files.ActionUpdate, path, &fileMetadata, logData); err != nil {
			writeError(w, buildErrors(err, "AuditError"), statusCode)
			return
		}

		if err := markFilePublished(ctx, path); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
