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

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/stretchr/testify/assert"
)

func TestContentItemUpdateSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

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

func TestContentItemUpdateWithBadBodyContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader("<json></json>"))

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestContentItemUpdateForbidden(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestContentItemUpdateForUnregisteredFile(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		return store.ErrFileNotRegistered
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, store.ErrFileNotRegistered
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestContentItemUpdateCreateFileEventError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return errors.New("failed to create file event")
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestContentItemUpdateReceivingUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "jan2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		return errors.New("unexpected error")
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestContentItemUpdateStoresPreviousSeriesIdWhenDatasetIDChanges(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "new_dataset_id", "edition": "jan2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	var capturedContentItem *files.StoredContentItem
	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		capturedContentItem = contentItem
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{
				DatasetID: "old_dataset_id",
				Edition:   "jan2026",
				Version:   "1",
			},
		}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"old_dataset_id"}, capturedContentItem.PreviousSeriesId)
	assert.Empty(t, capturedContentItem.PreviousEditionId)
}

func TestContentItemUpdateStoresPreviousEditionIdWhenEditionChanges(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "test_dataset_id", "edition": "feb2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	var capturedContentItem *files.StoredContentItem
	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		capturedContentItem = contentItem
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{
				DatasetID: "test_dataset_id",
				Edition:   "jan2026",
				Version:   "1",
			},
		}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, capturedContentItem.PreviousSeriesId)
	assert.Equal(t, []string{"jan2026"}, capturedContentItem.PreviousEditionId)
}

func TestContentItemUpdateAppendsPreviousIdsToExistingHistory(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "new_dataset_id", "edition": "mar2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	var capturedContentItem *files.StoredContentItem
	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		capturedContentItem = contentItem
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{
				DatasetID:         "current_dataset_id",
				Edition:           "feb2026",
				Version:           "1",
				PreviousSeriesId:  []string{"old_dataset_id"},
				PreviousEditionId: []string{"jan2026"},
			},
		}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"old_dataset_id", "current_dataset_id"}, capturedContentItem.PreviousSeriesId)
	assert.Equal(t, []string{"jan2026", "feb2026"}, capturedContentItem.PreviousEditionId)
}

func TestContentItemUpdateStoresBothPreviousIdsWhenDatasetIDAndEditionChange(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/files/file.txt", strings.NewReader(`{"content_item": {"dataset_id": "new_dataset_id", "edition": "feb2026", "version": "1"}}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	var capturedContentItem *files.StoredContentItem
	updateContentItemFunc := func(ctx context.Context, path string, contentItem *files.StoredContentItem) error {
		capturedContentItem = contentItem
		return nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{
				DatasetID: "old_dataset_id",
				Edition:   "jan2026",
				Version:   "1",
			},
		}, nil
	}

	h := api.HandlerUpdateContentItem(updateContentItemFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"old_dataset_id"}, capturedContentItem.PreviousSeriesId)
	assert.Equal(t, []string{"jan2026"}, capturedContentItem.PreviousEditionId)
}

