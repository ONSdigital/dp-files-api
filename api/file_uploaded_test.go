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

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestMarkFileUploadCompleteUnsuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"etag": "1234-asdfg-54321-qwerty"}`)
	req := httptest.NewRequest(http.MethodPatch, "/files/meme.jpg", body)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkUploadComplete(
		func(ctx context.Context, metaData files.FileEtagChange) error {
			return errors.New("it's all gone very wrong")
		},
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "it's all gone very wrong")
}

func TestJsonDecodingUploadComplete(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"etag": 1234,}`)
	req := httptest.NewRequest(http.MethodPatch, "/files/path.txt", body)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkUploadComplete(
		func(ctx context.Context, metaData files.FileEtagChange) error {
			return errors.New("it's all gone very wrong")
		},
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{}, nil
		},
		authMock,
		identityClientMock,
	)

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
			req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

			rec := httptest.NewRecorder()

			authMock, identityClientMock, _ := setUpAuthServices()

			h := api.HandleMarkUploadComplete(
				func(ctx context.Context, metaData files.FileEtagChange) error { return nil },
				func(ctx context.Context, event *files.FileEvent) error { return nil },
				func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
					return files.StoredRegisteredMetaData{}, nil
				},
				authMock,
				identityClientMock,
			)

			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			expectedResponse := fmt.Sprintf(`{"errors": [{"errorCode": "ValidationError", "description": %q}]}`, test.expectedErrorDescription)
			response, _ := io.ReadAll(rec.Body)
			assert.JSONEq(t, expectedResponse, string(response))
		})
	}
}

func TestMarkUploadComplete_AuditRecordCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"etag": "1234-asdfg-54321-qwerty"}`)
	req := httptest.NewRequest(http.MethodPatch, "/files/meme.jpg", body)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	auditEventCreated := false
	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkUploadComplete(
		func(ctx context.Context, metaData files.FileEtagChange) error { return nil },
		func(ctx context.Context, event *files.FileEvent) error {
			auditEventCreated = true
			assert.Equal(t, files.ActionUpdate, event.Action)
			assert.Equal(t, "admin", event.RequestedBy.ID)
			assert.NotNil(t, event.File)
			return nil
		},
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{Path: path}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, auditEventCreated)
}

func TestMarkUploadComplete_AuditRecordFailure_StillReturns200(t *testing.T) {
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"etag": "1234-asdfg-54321-qwerty"}`)
	req := httptest.NewRequest(http.MethodPatch, "/files/meme.jpg", body)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkUploadComplete(
		func(ctx context.Context, metaData files.FileEtagChange) error { return nil },
		func(ctx context.Context, event *files.FileEvent) error {
			return errors.New("failed to create audit record")
		},
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{Path: path}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
