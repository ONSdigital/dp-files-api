package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

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
	response, _ := io.ReadAll(rec.Body)
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
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "BadJsonEncoding")
}

func TestValidateUploadComplete(t *testing.T) {
	tests := []struct {
		name                     string
		incomingJSON             string
		expectedErrorDescription string
	}{
		{
			name:                     "Validate that etag is required",
			incomingJSON:             `{"": ""}`,
			expectedErrorDescription: "Etag required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := bytes.NewBufferString(test.incomingJSON)
			req := httptest.NewRequest(http.MethodPatch, "/files/path.jpg", body)

			rec := httptest.NewRecorder()

			nilFunc := func(ctx context.Context, metaData files.FileEtagChange) error { return nil }

			h := api.HandleMarkUploadComplete(nilFunc)
			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			expectedResponse := fmt.Sprintf(`{"errors": [{"code": "ValidationError", "description": %q}]}`, test.expectedErrorDescription)
			response, _ := io.ReadAll(rec.Body)
			assert.JSONEq(t, expectedResponse, string(response))
		})
	}
}
