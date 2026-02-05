package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/stretchr/testify/assert"
)

func TestContentItemUpdateWithBadBodyContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader("<json></json>"))

	h := api.HandlerUpdateContentItem(func(ctx context.Context, path string, contentItem files.StoredContentItem) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestContentItemUpdateForUnregisteredFile(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))

	h := api.HandlerUpdateContentItem(func(ctx context.Context, path string, contentItem files.StoredContentItem) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, store.ErrFileNotRegistered
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestContentItemUpdateReceivingUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))

	h := api.HandlerUpdateContentItem(func(ctx context.Context, path string, contentItem files.StoredContentItem) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, errors.New("unexpected error")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestContentItemUpdateSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))

	h := api.HandlerUpdateContentItem(func(ctx context.Context, path string, contentItem files.StoredContentItem) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{
				DatasetID: contentItem.DatasetID,
				Edition:   contentItem.Edition,
				Version:   contentItem.Version,
			}}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	expectedResponse, _ := json.Marshal(files.StoredRegisteredMetaData{
		ContentItem: &files.StoredContentItem{
			DatasetID: "test_dataset_id",
			Edition:   "jan2026",
			Version:   "1",
		},
	})
	response, _ := io.ReadAll(rec.Body)
	assert.JSONEq(t, string(expectedResponse), string(response))
}
