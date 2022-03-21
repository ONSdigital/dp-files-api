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
type MarkFilePublished func(ctx context.Context, path string) error
type MarkDecryptionComplete func(ctx context.Context, change files.FileEtagChange) error
type UpdateCollectionID func(ctx context.Context, path, collectionID string) error
type GetFilesMetadata func(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error)

func PatchRequestToHandler(uploadCompleteHandler http.HandlerFunc, publishedHandler http.HandlerFunc, decryptedHandler http.HandlerFunc, collectionUpdateHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		stateMetaData, err := getStateMetadataFromRequest(requestBody)

		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		// Reset request body after ReadAll for subsequent handlers
		setRequestBody(req, requestBody)

		if stateMetaData.CollectionID != nil && stateMetaData.State == nil {
			collectionUpdateHandler.ServeHTTP(w, req)
			return
		}

		switch *stateMetaData.State {
		case files.StateUploaded:
			uploadCompleteHandler.ServeHTTP(w, req)
		case files.StatePublished:
			publishedHandler.ServeHTTP(w, req)
		case files.StateDecrypted:
			decryptedHandler.ServeHTTP(w, req)
		default:
			log.Error(req.Context(), "InvalidStateChange", errors.New("invalid STATE change"), log.Data{"state": *stateMetaData.State})
			writeError(w, buildErrors(errors.New("invalid STATE change"), "InvalidStateChange"), http.StatusBadRequest)
		}
	}
}

func setRequestBody(req *http.Request, requestBody []byte) {
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
}

func getStateMetadataFromRequest(requestBody []byte) (StateMetadata, error) {
	stateMetaData := StateMetadata{}

	requestBodyBuffer := bytes.NewBuffer(requestBody)

	state := ioutil.NopCloser(requestBodyBuffer)

	err := json.NewDecoder(state).Decode(&stateMetaData)
	return stateMetaData, err
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

func HandleMarkFilePublished(markFilePublished MarkFilePublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		path := mux.Vars(req)["path"]

		if err := markFilePublished(req.Context(), path); err != nil {
			handleError(w, err)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleMarkCollectionPublished(markCollectionPublished MarkCollectionPublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		collectionID := mux.Vars(req)["collectionID"]

		if err := markCollectionPublished(req.Context(), collectionID); err != nil {
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

type FilesCollection struct {
	Count      int64                            `json:"count"`
	Limit      int64                            `json:"limit"`
	Offset     int64                            `json:"offset"`
	TotalCount int64                            `json:"total_count"`
	Items      []files.StoredRegisteredMetaData `json:"items"`
}

func HandlerGetFilesMetadata(getFilesMetadata GetFilesMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		collectionID := req.URL.Query().Get("collection_id")

		if collectionID == "" {
			err := errors.New("missing collection ID")
			writeError(w, buildErrors(err, "BadRequest"), http.StatusBadRequest)
			return
		}

		fm, err := getFilesMetadata(req.Context(), collectionID)
		if err != nil {
			handleError(w, err)
			return
		}

		fc := filesCollectionFromMetadata(fm)
		err = respondWithFilesCollectionJSON(w, fc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

	}
}

func filesCollectionFromMetadata(f []files.StoredRegisteredMetaData) FilesCollection {
	count := int64(len(f))
	fc := FilesCollection{
		Count:      count,
		Limit:      count,
		Offset:     0,
		TotalCount: count,
		Items:      f,
	}
	return fc
}

func HandlerRegisterUploadStarted(register RegisterFileUpload) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rm, err := getRegisterMetadataFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		err = validateRegisterMetadata(rm)
		if err != nil {
			handleError(w, err)
			return
		}

		err = register(req.Context(), generateStoredRegisterMetaData(rm))
		if err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func HandleMarkUploadComplete(markUploaded MarkUploadComplete) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ec, err := getEtagChangeFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err = validator.New().Struct(ec); err != nil {
			handleError(w, err)
			return
		}

		err = markUploaded(req.Context(), generateStoredUploadMetaData(ec, mux.Vars(req)["path"]))
		if err != nil {
			handleError(w, err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func respondWithFilesCollectionJSON(w http.ResponseWriter, fc FilesCollection) error {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(fc)
	return err
}

func validateRegisterMetadata(rm RegisterMetadata) error {
	validate := validator.New()
	validate.RegisterValidation("aws-upload-key", awsUploadKeyValidator)
	err := validate.Struct(rm)
	return err
}

func getRegisterMetadataFromRequest(req *http.Request) (RegisterMetadata, error) {
	rm := RegisterMetadata{}
	err := json.NewDecoder(req.Body).Decode(&rm)
	return rm, err
}

func getEtagChangeFromRequest(req *http.Request) (EtagChange, error) {
	ec := EtagChange{}
	err := json.NewDecoder(req.Body).Decode(&ec)
	return ec, err
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
