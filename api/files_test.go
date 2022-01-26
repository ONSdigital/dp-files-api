package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.CreateFileUploadStartedHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
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
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.CreateFileUploadStartedHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "BadJsonEncoding")
}

func TestValidationMetaDataCreation(t *testing.T) {
	tests := []struct {
		name                     string
		incomingJson             string
		expectedErrorDescription string
	}{
		{
			name:                     "Validate that path is required",
			incomingJson:             `{"is_publishable":true,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Path required",
		},
		{
			name:                     "Validate that path uri is valid",
			incomingJson:             `{"path": "/bad/path.jpg", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Path aws-upload-key",
		},
		{
			name:                     "Validate that is_publishable required",
			incomingJson:             `{"path": "some/file.txt", "collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "IsPublishable required",
		},
		{
			name:                     "Validate that collection_id is required",
			incomingJson:             `{"path": "some/file.txt", "is_publishable":false,"title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "CollectionID required",
		},
		{
			name:                     "Validate that size_in_bytes is positive integer",
			incomingJson:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 0, "type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "SizeInBytes gt",
		},
		{
			name:                     "Validate that type is valid mime type",
			incomingJson:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 10, "type":"image/jpeg123","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Type mime-type",
		},
		{
			name:                     "Validate that licence is required",
			incomingJson:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 10, "type":"image/jpeg","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "Licence required",
		},
		{
			name:                     "Validate that licence_url is required",
			incomingJson:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 10, "type":"image/jpeg","licence":"OGL v3"}`,
			expectedErrorDescription: "LicenceUrl required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := bytes.NewBufferString(test.incomingJson)
			req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

			rec := httptest.NewRecorder()

			errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error { return nil }

			h := api.CreateFileUploadStartedHandler(errFunc)
			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			expectedResponse := fmt.Sprintf(`{"errors": [{"code": "ValidationError", "description": "%s"}]}`, test.expectedErrorDescription)
			response, _ := ioutil.ReadAll(rec.Body)
			assert.JSONEq(t, expectedResponse, string(response))
		})
	}
}

func TestMarkFileUploadCompleteUnsuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "path": "images/meme.jpg",
          "etag": "1234-asdfg-54321-qwerty"
        }`)
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredUploadCompleteMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.CreateMarkUploadCompleteHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
}

func TestJsonDecodingUploadComplete(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "path": true,
          "etag": 1234,
        }`)
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredUploadCompleteMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.CreateMarkUploadCompleteHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "BadJsonEncoding")
}

func TestValidateUploadComplete(t *testing.T) {
	tests := []struct {
		name                     string
		incomingJson             string
		expectedErrorDescription string
	}{
		{
			name:                     "Validate that path is required",
			incomingJson:             `{"etag": "1234-asdfg-54321-qwerty"}`,
			expectedErrorDescription: "Path required",
		},
		{
			name:                     "Validate that path uri is valid",
			incomingJson:             `{"path": "/bad/path.jpg","etag": "1234-asdfg-54321-qwerty"}`,
			expectedErrorDescription: "Path aws-upload-key",
		},
		{
			name:                     "Validate that etag is required",
			incomingJson:             `{"path": "valid/path.jpg"}`,
			expectedErrorDescription: "Etag required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := bytes.NewBufferString(test.incomingJson)
			req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

			rec := httptest.NewRecorder()

			nilFunc := func(ctx context.Context, metaData files.StoredUploadCompleteMetaData) error { return nil }

			h := api.CreateMarkUploadCompleteHandler(nilFunc)
			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			expectedResponse := fmt.Sprintf(`{"errors": [{"code": "ValidationError", "description": "%s"}]}`, test.expectedErrorDescription)
			response, _ := ioutil.ReadAll(rec.Body)
			assert.JSONEq(t, expectedResponse, string(response))
		})
	}
}

func TestGetFileMetadataHandlesUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/files/file/path.jpg", nil)
	h := api.CreateGetFileMetadataHandler(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, errors.New("broken")
	})
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}
