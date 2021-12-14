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
	body := bytes.NewBufferString(`{"test": "test"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/files", body)

	errFunc := func (ctx context.Context, metaData files.MetaData) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.CreateFileUploadStartedHandler(errFunc)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
}
