package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestGetFilesMetadataWhenCollectionIDNotProvided(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error) {
		return []files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFilesMetadataWhenErrorReturned(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files?collection_id=12345678", nil)

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error) {
		return []files.StoredRegisteredMetaData{}, errors.New("something went wrong")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetFilesMetadataHandledWriteFailures(t *testing.T) {
	rec := &ErrorWriter{}
	req := httptest.NewRequest(http.MethodGet, "/files?collection_id=12345678", nil)

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error) {
		return []files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.status)
}
