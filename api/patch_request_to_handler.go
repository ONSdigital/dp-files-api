package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
	"io/ioutil"
	"net/http"
)

type PatchRequestHandlers struct {
	UploadComplete   http.HandlerFunc
	Published        http.HandlerFunc
	Decrypted        http.HandlerFunc
	CollectionUpdate http.HandlerFunc
}

func PatchRequestToHandler(handlers PatchRequestHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		stateMetaData, err := getStateMetadataFromRequest(req)

		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if stateMetaData.CollectionID != nil && stateMetaData.State == nil {
			handlers.CollectionUpdate.ServeHTTP(w, req)
			return
		}

		switch *stateMetaData.State {
		case files.StateUploaded:
			handlers.UploadComplete.ServeHTTP(w, req)
		case files.StatePublished:
			handlers.Published.ServeHTTP(w, req)
		case files.StateDecrypted:
			handlers.Decrypted.ServeHTTP(w, req)
		default:
			log.Error(req.Context(), "InvalidStateChange", errors.New("invalid state change"), log.Data{"state": *stateMetaData.State})
			writeError(w, buildErrors(errors.New("invalid state change"), "InvalidStateChange"), http.StatusBadRequest)
		}
	}
}

func getStateMetadataFromRequest(req *http.Request) (StateMetadata, error) {
	stateMetaData := StateMetadata{}
	requestBody, err := ioutil.ReadAll(req.Body)

	setRequestBody(req, requestBody)

	if err == nil {
		requestBodyBuffer := bytes.NewBuffer(requestBody)
		state := ioutil.NopCloser(requestBodyBuffer)
		err = json.NewDecoder(state).Decode(&stateMetaData)
	}

	return stateMetaData, err
}

func setRequestBody(req *http.Request, requestBody []byte) {
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
}
