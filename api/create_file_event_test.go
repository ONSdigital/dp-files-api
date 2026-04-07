package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/stretchr/testify/assert"
)

func setUpAuthServices() (authorisation.Middleware, *clientsidentity.Client, authorisation.PermissionsChecker) {
	cfg, _ := config.Get()

	testIdentityClient := clientsidentity.New(cfg.ZebedeeURL)

	permissionsChecker := &authMock.PermissionsCheckerMock{
		HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
			return true, nil
		},
	}

	authorisationMock := &authMock.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
		RequireWithAttributesFunc: func(permission string, handlerFunc http.HandlerFunc, getAttributes authorisation.GetAttributesFromRequest) http.HandlerFunc {
			return handlerFunc
		},
		ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			return &permissionsAPISDK.EntityData{UserID: "admin"}, nil
		},
	}
	return authorisationMock, testIdentityClient, permissionsChecker
}

func TestCreateFileEventWithBadJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader("<json></json>"))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()
	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error { return nil }, authMiddlewareMock, identityClientMock, permissionsMock)

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

	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return errors.New("database error")
	}, authMiddlewareMock, identityClientMock, permissionsMock)

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
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestCreateFileEventForbidden(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123", "email": "user@example.com"},
		"action": "READ",
		"resource": "/downloads/file.csv",
		"file": {"path": "file.csv", "type": "csv"}
	}`))
	req.Header.Add("Authorization", "fake-token")

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreateFileEventUnauthorised(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{
		"requested_by": {"id": "user123", "email": "user@example.com"},
		"action": "READ",
		"resource": "/downloads/file.csv",
		"file": {"path": "file.csv", "type": "csv"}
	}`))

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateFileEventWithEmptyJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file-events", strings.NewReader(`{}`))
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

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
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

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
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

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
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

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
	req.Header.Add("Authorization", authorisationtest.AdminJWTToken)

	authMiddlewareMock, identityClientMock, permissionsMock := setUpAuthServices()

	h := api.HandlerCreateFileEvent(func(ctx context.Context, event *files.FileEvent) error {
		return nil
	}, authMiddlewareMock, identityClientMock, permissionsMock)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
