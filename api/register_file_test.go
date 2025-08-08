package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestFileMetaDataCreationUnsuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "path": "images/meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }`)
	req := httptest.NewRequest(http.MethodPost, "/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.HandlerRegisterUploadStarted(errFunc, 5*time.Second)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
}

func TestFileMetaDataCreationUnsuccessfulWithBothCollectionAndBundleID(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "path": "images/meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
		  "bundle_id": "test-bundle",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }`)
	req := httptest.NewRequest(http.MethodPost, "/files", body)

	h := api.HandlerRegisterUploadStarted(func(ctx context.Context, metaData files.StoredRegisteredMetaData) error { return nil }, 5*time.Second)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "cannot set both collection and bundle ID")
}

func TestJsonDecodingMetaDataCreation(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "path": 123,
          "is_publishable": "true",
          "collection_id": false,
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }`)
	req := httptest.NewRequest(http.MethodPost, "/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.HandlerRegisterUploadStarted(errFunc, 5*time.Second)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "BadJsonEncoding")
}

func TestValidationMetaDataCreation(t *testing.T) {
	tests := []struct {
		name                     string
		incomingJSON             string
		expectedErrorDescription string
	}{
		{
			name:                     "Validate that path is required",
			incomingJSON:             `{"is_publishable":true,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Path required",
		},
		{
			name:                     "Validate that path uri is valid",
			incomingJSON:             `{"path": "/bad/path.jpg", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Path aws-upload-key",
		},
		{
			name:                     "Validate that is_publishable required",
			incomingJSON:             `{"path": "some/file.txt", "collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "IsPublishable required",
		},
		{
			name:                     "Validate that size_in_bytes is positive integer",
			incomingJSON:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 0, "type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "SizeInBytes gt",
		},
		{
			name:                     "Validate that licence is required",
			incomingJSON:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 10, "type":"image/jpeg","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Licence required",
		},
		{
			name:                     "Validate that licence_url is required",
			incomingJSON:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 10, "type":"image/jpeg","licence":"OGL v3"}`,
			expectedErrorDescription: "LicenceURL required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := bytes.NewBufferString(test.incomingJSON)
			req := httptest.NewRequest(http.MethodPost, "/files", body)

			rec := httptest.NewRecorder()

			errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error { return nil }

			h := api.HandlerRegisterUploadStarted(errFunc, 5*time.Second)
			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			expectedResponse := fmt.Sprintf(`{"errors": [{"errorCode": "ValidationError", "description": %q}]}`, test.expectedErrorDescription)
			response, _ := io.ReadAll(rec.Body)
			assert.JSONEq(t, expectedResponse, string(response))
		})
	}
}

func TestCollectionIDOmittedFromBodyDoesNotRaiseError(t *testing.T) {
	body := `{"path": "some/file.txt", "is_publishable":false,"title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(body))

	h := api.HandlerRegisterUploadStarted(func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		assert.Nil(t, metaData.CollectionID)
		return nil
	}, 5*time.Second)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCollectionIDInBodyDoesNotRaiseError(t *testing.T) {
	collectionID := "1234"
	body := fmt.Sprintf(`{"path": "some/file.txt", "collection_id": %q, "is_publishable":false,"title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`, collectionID)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(body))

	h := api.HandlerRegisterUploadStarted(func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		assert.Equal(t, collectionID, *metaData.CollectionID)
		return nil
	}, 5*time.Second)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}
