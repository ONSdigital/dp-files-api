package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/go-playground/validator"
	"net/http"
)

type ApiMetaData struct {
	Path          string    `json:"path" validate:"required"`
	IsPublishable bool      `json:"is_publishable"`
	CollectionID  string    `json:"collection_id"`
	Title         string    `json:"title" `
	SizeInBytes   int64     `json:"size_in_bytes"`
	Type          string    `json:"type"`
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
		err = validate.Struct(m)
		if err != nil {
			handleError(w, err)
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
		IsPublishable: m.IsPublishable,
		CollectionID:  m.CollectionID,
		Title:         m.Title,
		SizeInBytes:   m.SizeInBytes,
		Type:          m.Type,
		Licence:       m.Licence,
		LicenceUrl:    m.LicenceUrl,
	}
}
