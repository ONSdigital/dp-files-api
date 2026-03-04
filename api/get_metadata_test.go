package api_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestGetFileMetadataHandlesUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, errors.New("broken")
	})
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}

func TestGetFileMetadataWithAuthSuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandleGetFileMetadataWithAuth(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{Path: "/files/path.jpg"}, nil
	}, authMock, identityClientMock, permissionsMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "/files/path.jpg")
}

func TestGetFileMetadataWithAuthUnauthorised(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)

	authMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandleGetFileMetadataWithAuth(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{Path: "/files/path.jpg"}, nil
	}, authMock, identityClientMock, permissionsMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "The user is unauthorised")
}

func TestGetFileMetadataWithAuthForbidden(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Add("Authorization", "test-invalid-token")

	authMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandleGetFileMetadataWithAuth(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{Path: "/files/path.jpg"}, nil
	}, authMock, identityClientMock, permissionsMock)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "the request was not authorised - check token and user's permissions")
}
