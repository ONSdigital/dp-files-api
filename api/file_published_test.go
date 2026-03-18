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
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/stretchr/testify/assert"
)

func TestMarkFilePublished_Successful(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFilePublished(
		func(ctx context.Context, path string) error { return nil },
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{Path: path}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMarkFilePublished_Unsuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFilePublished(
		func(ctx context.Context, path string) error {
			return errors.New("it's all gone very wrong")
		},
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMarkFilePublished_FileNotRegistered(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFilePublished(
		func(ctx context.Context, path string) error {
			return store.ErrFileNotRegistered
		},
		func(ctx context.Context, event *files.FileEvent) error { return nil },
		func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
			return files.StoredRegisteredMetaData{}, nil
		},
		authMock,
		identityClientMock,
	)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMarkFilePublished_AuditRecordCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	auditEventCreated := false
	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFilePublished(
		func(ctx context.Context, path string) error { return nil },
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

func TestMarkFilePublished_AuditRecordFailure_StillReturns200(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/file.txt", http.NoBody)
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMock, identityClientMock, _ := setUpAuthServices()

	h := api.HandleMarkFilePublished(
		func(ctx context.Context, path string) error { return nil },
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

	assert.Equal(t, http.StatusOK, rec.Code)
}
