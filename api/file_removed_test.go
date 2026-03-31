package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestHandleRemoveFile_Successful(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	removeFileFunc := func(ctx context.Context, path string, fileMetadata files.StoredRegisteredMetaData) error {
		return nil
	}

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleRemoveFile(removeFileFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandleRemoveFile_Forbidden(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", http.NoBody)

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleRemoveFile(nil, nil, nil, authMiddlewareMock, identityClientMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleRemoveFile_GetFileMetadataError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, errors.New("failed to get metadata")
	}

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleRemoveFile(nil, nil, getFileMetadataFunc, authMiddlewareMock, identityClientMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleRemoveFile_CreateFileEventError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return errors.New("failed to create file event")
	}

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleRemoveFile(nil, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleRemoveFile_RemoveFileError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	getFileMetadataFunc := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}
	createFileEventFunc := func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}
	removeFileFunc := func(ctx context.Context, path string, fileMetadata files.StoredRegisteredMetaData) error {
		return errors.New("failed to remove file")
	}

	authMiddlewareMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleRemoveFile(removeFileFunc, createFileEventFunc, getFileMetadataFunc, authMiddlewareMock, identityClientMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
