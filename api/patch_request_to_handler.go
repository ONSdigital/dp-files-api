package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/log.go/v2/log"
)

type PatchRequestHandlers struct {
	UploadComplete   http.HandlerFunc
	Published        http.HandlerFunc
	Moved            http.HandlerFunc
	CollectionUpdate http.HandlerFunc
}

type StateMetadata struct {
	State        *string `json:"state,omitempty"`
	CollectionID *string `json:"collection_id,omitempty"`
}

func PatchRequestToHandler(handlers PatchRequestHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		stateMetaData, err := getStateMetadataFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if stateMetaData.CollectionID != nil && isCollectionIDUpdate(stateMetaData) {
			handlers.CollectionUpdate.ServeHTTP(w, req)
			return
		}

		switch *stateMetaData.State {
		case store.StateUploaded:
			handlers.UploadComplete.ServeHTTP(w, req)
		case store.StatePublished:
			handlers.Published.ServeHTTP(w, req)
		case store.StateMoved:
			handlers.Moved.ServeHTTP(w, req)
		default:
			log.Error(req.Context(), "InvalidStateChange", errors.New("invalid state change"), log.Data{"state": *stateMetaData.State})
			writeError(w, buildErrors(errors.New("invalid state change"), "InvalidStateChange"), http.StatusBadRequest)
		}
	}
}

func isCollectionIDUpdate(stateMetaData StateMetadata) bool {
	return stateMetaData.State == nil
}

func getStateMetadataFromRequest(req *http.Request) (StateMetadata, error) {
	stateMetaData := StateMetadata{}
	requestBody, err := io.ReadAll(req.Body)

	setRequestBody(req, requestBody)

	if err == nil {
		requestBodyBuffer := bytes.NewBuffer(requestBody)
		state := io.NopCloser(requestBodyBuffer)
		err = json.NewDecoder(state).Decode(&stateMetaData)
	}

	return stateMetaData, err
}

func setRequestBody(req *http.Request, requestBody []byte) {
	req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
}
