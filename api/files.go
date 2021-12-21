package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/go-playground/validator"
	"net/http"
)

type RegisterMetaData struct {
	Path          string `json:"path" validate:"required,uri"`
	IsPublishable *bool  `json:"is_publishable,omitempty" validate:"required"`
	CollectionID  string `json:"collection_id" validate:"required"`
	Title         string `json:"title"`
	SizeInBytes   uint64 `json:"size_in_bytes" validate:"gt=0"`
	Type          string `json:"type" validate:"mime-type"`
	Licence       string `json:"licence" validate:"required"`
	LicenceUrl    string `json:"licence_url" validate:"required"`
}

type UploadCompleteMetaData struct {
	Path string `json:"path" validate:"required,uri"`
	Etag string `json:"etag" validate:"required"`
}

func CreateFileUploadStartedHandler(register files.CreateUploadStartedEntry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		m := RegisterMetaData{}

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		validate := validator.New()
		validate.RegisterValidation("mime-type", mimeValidator)
		err = validate.Struct(m)
		if err != nil {
			handleError(w, err)
			return
		}

		err = register(req.Context(), generateStoredRegisterMetaData(m))
		if err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func MarkUploadCompleteHandler(markUploaded files.MarkUploadComplete) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m := UploadCompleteMetaData{}

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		err = validator.New().Struct(m)
		if err != nil {
			handleError(w, err)
			return
		}

		err = markUploaded(req.Context(), generateStoredUploadMetaData(m))
		if err != nil {
			handleError(w, err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func generateStoredRegisterMetaData(m RegisterMetaData) files.StoredRegisteredMetaData {
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

func generateStoredUploadMetaData(m UploadCompleteMetaData) files.StoredUploadCompleteMetaData {
	return files.StoredUploadCompleteMetaData{
		Path: m.Path,
		Etag: m.Etag,
	}
}
