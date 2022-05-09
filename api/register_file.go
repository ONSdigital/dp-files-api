package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator"

	"github.com/ONSdigital/dp-files-api/files"
)

type RegisterFileUpload func(ctx context.Context, metaData files.StoredRegisteredMetaData) error

type RegisterMetadata struct {
	Path          string  `json:"path" validate:"required,aws-upload-key"`
	IsPublishable *bool   `json:"is_publishable,omitempty" validate:"required"`
	CollectionID  *string `json:"collection_id,omitempty"`
	Title         string  `json:"title"`
	SizeInBytes   uint64  `json:"size_in_bytes" validate:"gt=0"`
	Type          string  `json:"type"`
	Licence       string  `json:"licence" validate:"required"`
	LicenceUrl    string  `json:"licence_url" validate:"required"`
}

func HandlerRegisterUploadStarted(register RegisterFileUpload) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rm, err := getRegisterMetadataFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := validateRegisterMetadata(rm); err != nil {
			handleError(w, err)
			return
		}

		if err := register(req.Context(), generateStoredRegisterMetaData(rm)); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func validateRegisterMetadata(rm RegisterMetadata) error {
	validate := validator.New()
	validate.RegisterValidation("aws-upload-key", awsUploadKeyValidator)
	return validate.Struct(rm)
}

func getRegisterMetadataFromRequest(req *http.Request) (RegisterMetadata, error) {
	rm := RegisterMetadata{}
	return rm, json.NewDecoder(req.Body).Decode(&rm)
}

func generateStoredRegisterMetaData(m RegisterMetadata) files.StoredRegisteredMetaData {
	return files.StoredRegisteredMetaData{
		Path:          m.Path,
		IsPublishable: *m.IsPublishable,
		CollectionID:  m.CollectionID,
		Title:         m.Title,
		SizeInBytes:   m.SizeInBytes,
		Type:          m.Type,
		Licence:       m.Licence,
		LicenceUrl:    m.LicenceUrl,
	}
}
