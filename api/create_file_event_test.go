package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestCreateFileEventWithBadJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader("<json></json>"))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFileEventWithStoreError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123"},
		"action": "READ",
		"resource": "/downloads/file.csv",
		"file": {"path": "file.csv", "type": "csv"}
	}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return errors.New("database error")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateFileEventSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123", "email": "user@example.com"},
		"action": "READ",
		"resource": "/downloads/file.csv",
		"file": {"path": "file.csv", "type": "csv"}
	}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestCreateFileEventWithEmptyJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFileEventWithMissingRequestedBy(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"action": "READ",
		"resource": "/downloads/file.csv",
		"file": {"path": "file.csv", "type": "csv"}
	}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFileEventWithMissingAction(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123"},
		"resource": "/downloads/file.csv",
		"file": {"path": "file.csv", "type": "csv"}
	}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFileEventWithMissingResource(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123"},
		"action": "READ",
		"file": {"path": "file.csv", "type": "csv"}
	}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFileEventWithMissingFilePath(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123"},
		"action": "READ",
		"resource": "/downloads/file.csv",
		"file": {"type": "csv"}
	}`))

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
