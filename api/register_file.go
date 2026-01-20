package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
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

func HandlerRegisterUploadStarted(register RegisterFileUpload, deadlineDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
		defer cancel()

		if err := register(ctx, generateStoredRegisterMetaData(rm)); err != nil {
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
