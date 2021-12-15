package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/go-playground/validator"
	"net/http"
)

type ApiMetaData struct {
	Path          string    `json:"path" validate:"required,uri"`
	IsPublishable *bool     `json:"is_publishable,omitempty" validate:"required"`
	CollectionID  string    `json:"collection_id" validate:"required"`
	Title         string    `json:"title"`
	SizeInBytes   uint64    `json:"size_in_bytes" validate:"gt=0"`
	Type          string    `json:"type" validate:"mime-type"`
	Licence       string    `json:"licence"`
	LicenceUrl    string    `json:"licence_url"`
	State         string    `json:"state"`
}

func CreateFileUploadStartedHandler(creatorFunc files.CreateUploadStartedEntry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		m := ApiMetaData{}

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			return
		}


		validate := validator.New()
		validate.RegisterValidation("mime-type", mimeValidator)
		err = validate.Struct(m)
		if err != nil {
			handleError(w, err)
			return
		}

		err = creatorFunc(req.Context(), generateStoredMetaData(m))
		if err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func generateStoredMetaData(m ApiMetaData) files.StoredMetaData {
	return files.StoredMetaData{
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
