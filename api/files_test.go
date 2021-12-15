package api_test

import (
	"bytes"
	"context"
	"errors"
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
          "path": "/images/meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }`)
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func (ctx context.Context, metaData files.StoredMetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.CreateFileUploadStartedHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
}

func TestPathIsValidFilePath(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }`)
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func (ctx context.Context, metaData files.StoredMetaData) error {
		return nil
	}

	h := api.CreateFileUploadStartedHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	expectedResponse := `
{
	"errors": [
		{
			"code": "ValidationError",
			"description": "Path required"
		}
	]
}
`
	response, _ := ioutil.ReadAll(rec.Body)
	assert.JSONEq(t, expectedResponse, string(response))
}
