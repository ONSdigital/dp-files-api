package api

import (
	"context"
	"encoding/json"
	"net/http"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"

	"github.com/go-playground/validator"
)

type MarkUploadComplete func(ctx context.Context, metaData files.FileEtagChange) error

func HandleMarkUploadComplete(markUploaded MarkUploadComplete, createFileEvent CreateFileEvent, getFileMetadata GetFileMetadata, authMiddleware auth.Middleware, idClient *clientsidentity.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		path := mux.Vars(req)["path"]

		logData := log.Data{
			"method": req.Method,
			"path":   path,
		}

		ec, err := getEtagChangeFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err = validator.New().Struct(ec); err != nil {
			handleError(w, err)
			return
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

		if err := markUploaded(ctx, generateFileEtagChange(ec, path)); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getEtagChangeFromRequest(req *http.Request) (EtagChange, error) {
	ec := EtagChange{}
	err := json.NewDecoder(req.Body).Decode(&ec)
	return ec, err
}
