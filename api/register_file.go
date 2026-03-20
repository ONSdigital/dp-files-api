package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

type RegisterFileUpload func(ctx context.Context, metaData files.StoredRegisteredMetaData) error

type RegisterMetadata struct {
	Path          string       `json:"path" validate:"required,aws-upload-key"`
	IsPublishable *bool        `json:"is_publishable,omitempty" validate:"required"`
	CollectionID  *string      `json:"collection_id,omitempty"`
	BundleID      *string      `json:"bundle_id,omitempty"`
	Title         string       `json:"title"`
	SizeInBytes   uint64       `json:"size_in_bytes" validate:"gt=0"`
	Type          string       `json:"type"`
	Licence       string       `json:"licence" validate:"required"`
	LicenceURL    string       `json:"licence_url" validate:"required"`
	ContentItem   *ContentItem `json:"content_item,omitempty"`
}

type ContentItem struct {
	DatasetID string `json:"dataset_id,omitempty"`
	Edition   string `json:"edition,omitempty"`
	Version   string `json:"version,omitempty"`
}

func HandlerRegisterUploadStarted(register RegisterFileUpload, createFileEvent CreateFileEvent, authMiddleware auth.Middleware, identityClient *clientsidentity.Client, deadlineDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), deadlineDuration)
		defer cancel()

		logData := log.Data{
			"method": req.Method,
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

		identityType := log.USER
		if authEntityData.IsServiceAuth {
			identityType = log.SERVICE
		}
		logAuth := log.Auth(identityType, authEntityData.EntityData.UserID)

		rm, err := getRegisterMetadataFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if rm.CollectionID != nil && rm.BundleID != nil {
			writeError(w, buildErrors(store.ErrBothCollectionAndBundleIDSet, "BothCollectionAndBundleIDSet"), http.StatusBadRequest)
			return
		}

		if err := validateRegisterMetadata(rm); err != nil {
			handleError(w, err)
			return
		}
		storedRegisterMetadata := generateStoredRegisterMetaData(rm)

		fileEvent := &files.FileEvent{
			RequestedBy: &files.RequestedBy{
				ID: authEntityData.EntityData.UserID,
			},
			Action:   files.ActionCreate,
			Resource: rm.Path,
			File:     &storedRegisterMetadata,
		}

		if err := createFileEvent(ctx, fileEvent); err != nil {
			log.Error(ctx, "failed to create file event", err, log.Classification(log.ProtectiveMonitoring), logAuth, logData)
			handleError(w, err)
			return
		}
		log.Info(ctx, "successfully created file event for file creation", log.Classification(log.ProtectiveMonitoring), logAuth, logData)

		if err := register(ctx, storedRegisterMetadata); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func validateRegisterMetadata(rm RegisterMetadata) error {
	validate := validator.New()
	if err := validate.RegisterValidation("aws-upload-key", awsUploadKeyValidator); err != nil {
		return err
	}
	return validate.Struct(rm)
}

func getRegisterMetadataFromRequest(req *http.Request) (RegisterMetadata, error) {
	rm := RegisterMetadata{}
	err := json.NewDecoder(req.Body).Decode(&rm)
	return rm, err
}

func generateStoredRegisterMetaData(m RegisterMetadata) files.StoredRegisteredMetaData {
	var contentItem *files.StoredContentItem
	if m.ContentItem != nil {
		contentItem = &files.StoredContentItem{
			DatasetID: m.ContentItem.DatasetID,
			Edition:   m.ContentItem.Edition,
			Version:   m.ContentItem.Version,
		}
	}

	return files.StoredRegisteredMetaData{
		Path:          m.Path,
		IsPublishable: *m.IsPublishable,
		CollectionID:  m.CollectionID,
		BundleID:      m.BundleID,
		Title:         m.Title,
		SizeInBytes:   m.SizeInBytes,
		Type:          m.Type,
		Licence:       m.Licence,
		LicenceURL:    m.LicenceURL,
		ContentItem:   contentItem,
	}
}
