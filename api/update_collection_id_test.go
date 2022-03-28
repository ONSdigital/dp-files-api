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

	h := api.HandlerUpdateCollectionID(func(ctx context.Context, path, collectionID string) error { return store.ErrFileNotRegistered })

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
