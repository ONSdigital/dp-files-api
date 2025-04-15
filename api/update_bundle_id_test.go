package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/stretchr/testify/assert"
)

func TestBundleIDUpdateWithBadBodyContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader("<json></json>"))

	h := api.HandlerUpdateBundleID(func(ctx context.Context, path, bundleID string) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBundleIDUpdateForUnregisteredFile(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"bundle_id": "123456789"}`))

	h := api.HandlerUpdateBundleID(func(ctx context.Context, path, bundleID string) error { return store.ErrFileNotRegistered })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBundleIDUpdateReceivingUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", strings.NewReader(`{"bundle_id": "123456789"}`))

	h := api.HandlerUpdateBundleID(func(ctx context.Context, path, bundleID string) error { return errors.New("broken") })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
