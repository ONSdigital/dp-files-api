package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	h := api.HandlerRegisterUploadStarted(errFunc)
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
	req := httptest.NewRequest(http.MethodPost, "/files", body)

	errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.HandlerRegisterUploadStarted(errFunc)
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
			name:                     "Validate that size_in_bytes is positive integer",
			incomingJson:             `{"path": "some/file.txt", "is_publishable":false,"collection_id":"1234-asdfg-54321-qwerty","title":"The latest Meme", "size_in_bytes": 0, "type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`,
			expectedErrorDescription: "SizeInBytes gt",
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
			req := httptest.NewRequest(http.MethodPost, "/files", body)

			rec := httptest.NewRecorder()

			errFunc := func(ctx context.Context, metaData files.StoredRegisteredMetaData) error { return nil }

			h := api.HandlerRegisterUploadStarted(errFunc)
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
          "etag": "1234-asdfg-54321-qwerty"
        }`)
	req := httptest.NewRequest(http.MethodPatch, "/files/meme.jpg", body)

	errFunc := func(ctx context.Context, metaData files.FileEtagChange) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.HandleMarkUploadComplete(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
}

func TestJsonDecodingUploadComplete(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "etag": 1234,
        }`)
	req := httptest.NewRequest(http.MethodPatch, "/files/path.txt", body)

	errFunc := func(ctx context.Context, metaData files.FileEtagChange) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.HandleMarkUploadComplete(errFunc)
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
			name:                     "Validate that etag is required",
			incomingJson:             `{"": ""}`,
			expectedErrorDescription: "Etag required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := bytes.NewBufferString(test.incomingJson)
			req := httptest.NewRequest(http.MethodPatch, "/files/path.jpg", body)

			rec := httptest.NewRecorder()

			nilFunc := func(ctx context.Context, metaData files.FileEtagChange) error { return nil }

			h := api.HandleMarkUploadComplete(nilFunc)
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
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", nil)
	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, errors.New("broken")
	})
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}

func TestPublishHandlerHandlesInvalidJSONContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/ignore.txt", strings.NewReader("<json>invalid</json>"))

	h := api.HandleMarkCollectionPublished(func(ctx context.Context, collectionID string) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "BadJsonEncoding")
}

func TestPublishHandlerHandlesUnexpectedPublishingError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/ignore.txt", strings.NewReader(`{"collection_id": "asdfghjkl"}`))

	h := api.HandleMarkCollectionPublished(func(ctx context.Context, collectionID string) error {
		return errors.New("broken")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}

func TestDecryptedHandlerHandlesInvalidJSONContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader("<json>invalid</json>"))

	h := api.HandleMarkFileDecrypted(func(ctx context.Context, change files.FileEtagChange) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "BadJsonEncoding")
}

func TestDecryptedHandlerHandlesUnexpectedPublishingError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"path": "dir/file.txt"}`))

	h := api.HandleMarkFileDecrypted(func(ctx context.Context, change files.FileEtagChange) error {
		return errors.New("broken")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}

func TestCollectionIDOmittedFromBodyDoesNotRaiseError(t *testing.T) {
	body := `{"path": "some/file.txt", "is_publishable":false,"title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(body))

	h := api.HandlerRegisterUploadStarted(func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		assert.Nil(t, metaData.CollectionID)
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCollectionIDInBodyDoesNotRaiseError(t *testing.T) {
	collectionID := "1234"
	body := fmt.Sprintf(`{"path": "some/file.txt", "collection_id": "%s", "is_publishable":false,"title":"The latest Meme","size_in_bytes":14794,"type":"image/jpeg","licence":"OGL v3","licence_url":"http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"}`, collectionID)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(body))

	h := api.HandlerRegisterUploadStarted(func(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
		assert.Equal(t, collectionID, *metaData.CollectionID)
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCollectionIDUpdateWithBadBodyContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader("<json></json>"))

	h := api.HandlerUpdateCollectionID(func(ctx context.Context, path, collectionID string) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCollectionIDUpdateForUnregisteredFile(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"collection_id": "123456789"}`))

	h := api.HandlerUpdateCollectionID(func(ctx context.Context, path, collectionID string) error { return files.ErrFileNotRegistered })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCollectionIDUpdateReceivingUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"collection_id": "123456789"}`))

	h := api.HandlerUpdateCollectionID(func(ctx context.Context, path, collectionID string) error { return errors.New("broken") })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
