package api_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestMovedHandlerHandlesInvalidJSONContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader("<json>invalid</json>"))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFileMoved(
		func(ctx context.Context, change files.FileEtagChange) error { return nil },
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

func TestMovedHandlerHandlesUnexpectedPublishingError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"etag": "abc123"}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFileMoved(
		func(ctx context.Context, change files.FileEtagChange) error { return errors.New("broken") },
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
	assert.Contains(t, string(response), "InternalError")
}

func TestMovedHandler_AuditRecordCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"etag": "abc123"}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	auditEventCreated := false
	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFileMoved(
		func(ctx context.Context, change files.FileEtagChange) error { return nil },
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

func TestMovedHandler_AuditRecordFailure_Returns500(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"etag": "abc123"}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFileMoved(
		func(ctx context.Context, change files.FileEtagChange) error { return nil },
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

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMovedHandler_GetFileMetadataError_Returns500(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"etag": "abc123"}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFileMoved(
		func(ctx context.Context, change files.FileEtagChange) error { return nil },
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{}, errors.New("failed to get file metadata")
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMovedHandler_NoToken_Returns401(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"etag": "abc123"}`))

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFileMoved(
		func(ctx context.Context, change files.FileEtagChange) error { return nil },
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{Path: "file.txt"}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
