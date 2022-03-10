package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

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

type StateMetadata struct {
	State        *string `json:"state,omitempty"`
	CollectionID *string `json:"collection_id,omitempty"`
}

type EtagChange struct {
	Etag string `json:"etag" validate:"required"`
}

type PublishData struct {
	CollectionID string `json:"collection_id"`
}

type RegisterFileUpload func(ctx context.Context, metaData files.StoredRegisteredMetaData) error
type MarkUploadComplete func(ctx context.Context, metaData files.FileEtagChange) error
type GetFileMetadata func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error)
type MarkCollectionPublished func(ctx context.Context, collectionID string) error
type MarkDecryptionComplete func(ctx context.Context, change files.FileEtagChange) error
type UpdateCollectionID func(ctx context.Context, path, collectionID string) error

func StateToHandler(uploadComplete http.HandlerFunc, published http.HandlerFunc, decrypted http.HandlerFunc, collectionUpdate http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		m := StateMetadata{}
		b, _ := ioutil.ReadAll(req.Body)

		state := ioutil.NopCloser(bytes.NewBuffer(b))
		if err := json.NewDecoder(state).Decode(&m); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		req.Body = ioutil.NopCloser(bytes.NewBuffer(b))

		if m.CollectionID != nil && m.State == nil {
			collectionUpdate.ServeHTTP(w, req)
			return
		}

		if *m.State == files.StateUploaded {
			uploadComplete.ServeHTTP(w, req)
		} else if *m.State == files.StatePublished {
			published.ServeHTTP(w, req)
		} else if *m.State == files.StateDecrypted {
			decrypted.ServeHTTP(w, req)
		} else {
			log.Error(req.Context(), "InvalidStateChange", errors.New("Invalid STATE change"), log.Data{"state": *m.State})
			writeError(w, buildErrors(errors.New("Invalid STATE change"), "InvalidStateChange"), http.StatusBadRequest)
		}
	}
}

type CollectionChange struct {
	CollectionID string `json:"collection_id"`
}

func HandlerUpdateCollectionID(updateCollectionID UpdateCollectionID) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cc := CollectionChange{}
		if err := json.NewDecoder(req.Body).Decode(&cc); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := updateCollectionID(req.Context(), mux.Vars(req)["path"], cc.CollectionID); err != nil {
			handleError(w, err)
		}
	}
}

func HandleMarkFileDecrypted(markDecryptionComplete MarkDecryptionComplete) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m := EtagChange{}
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := markDecryptionComplete(req.Context(), generateStoredUploadMetaData(m, mux.Vars(req)["path"])); err != nil {
			handleError(w, err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func HandleMarkCollectionPublished(markCollectionPublished MarkCollectionPublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		p := PublishData{}
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := markCollectionPublished(req.Context(), p.CollectionID); err != nil {
			handleError(w, err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func HandleGetFileMetadata(getMetadata GetFileMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")
		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}

		json.NewEncoder(w).Encode(metadata)

		w.WriteHeader(http.StatusOK)
	}
}

func HandlerRegisterUploadStarted(register RegisterFileUpload) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		m := RegisterMetadata{}
		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		validate := validator.New()
		validate.RegisterValidation("aws-upload-key", awsUploadKeyValidator)
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

func HandleMarkUploadComplete(markUploaded MarkUploadComplete) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m := EtagChange{}
		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err = validator.New().Struct(m); err != nil {
			handleError(w, err)
			return
		}

		err = markUploaded(req.Context(), generateStoredUploadMetaData(m, mux.Vars(req)["path"]))
		if err != nil {
			handleError(w, err)
		}

		w.WriteHeader(http.StatusOK)
	}
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

func generateStoredUploadMetaData(m EtagChange, path string) files.FileEtagChange {
	return files.FileEtagChange{
		Path: path,
		Etag: m.Etag,
	}
}
