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

func TestGetFilesMetadataWhenNoIDsProvided(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files", http.NoBody)

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID, bundleID string) ([]files.StoredRegisteredMetaData, error) {
		return []files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFilesMetadataWithCollectionID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files?collection_id=collection1", http.NoBody)

	calledWithCollection := ""
	calledWithBundle := ""

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID, bundleID string) ([]files.StoredRegisteredMetaData, error) {
		calledWithCollection = collectionID
		calledWithBundle = bundleID
		return []files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "collection1", calledWithCollection)
	assert.Equal(t, "", calledWithBundle)
}

func TestGetFilesMetadataWithBundleID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files?bundle_id=bundle1", http.NoBody)

	calledWithCollection := ""
	calledWithBundle := ""

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID, bundleID string) ([]files.StoredRegisteredMetaData, error) {
		calledWithCollection = collectionID
		calledWithBundle = bundleID
		return []files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", calledWithCollection)
	assert.Equal(t, "bundle1", calledWithBundle)
}

func TestGetFilesMetadataWhenErrorReturned(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files?collection_id=12345678", http.NoBody)

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID, bundleID string) ([]files.StoredRegisteredMetaData, error) {
		return []files.StoredRegisteredMetaData{}, errors.New("something went wrong")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetFilesMetadataHandledWriteFailures(t *testing.T) {
	rec := &ErrorWriter{}
	req := httptest.NewRequest(http.MethodGet, "/files?collection_id=12345678", http.NoBody)

	h := api.HandlerGetFilesMetadata(func(ctx context.Context, collectionID, bundleID string) ([]files.StoredRegisteredMetaData, error) {
		return []files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.status)
}
