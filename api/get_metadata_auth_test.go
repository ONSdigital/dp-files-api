package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/stretchr/testify/assert"
)

func TestGetFileMetadataAuth_UsesDatasetEditionAttributesAfterMetadataLoaded(t *testing.T) {
	metadataLoaded := false
	metadata := files.StoredRegisteredMetaData{
		Path: "path.jpg",
		ContentItem: &files.StoredContentItem{
			DatasetID: "dataset-1",
			Edition:   "edition-1",
			Version:   "1",
		},
	}

	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		metadataLoaded = true
		return metadata, nil
	}

	permissionsChecker := &authMock.PermissionsCheckerMock{
		HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
			if !metadataLoaded {
				t.Fatalf("permissions check happened before metadata was loaded")
			}
			assert.Equal(t, "static-files:read", permission)
			assert.Equal(t, map[string]string{"dataset_edition": "dataset-1/edition-1"}, attributes)
			return true, nil
		},
	}

	middleware := &authMock.MiddlewareMock{
		ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			return &permissionsAPISDK.EntityData{UserID: "user-1"}, nil
		},
	}

	handler := api.HandleGetFileMetadataWithPermissions(getMetadata, middleware, permissionsChecker, &authMock.ZebedeeClientMock{})

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-valid-jwt-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileMetadataAuth_MissingJWTReturns401(t *testing.T) {
	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	handler := api.HandleGetFileMetadataWithPermissions(
		getMetadata,
		&authMock.MiddlewareMock{ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			t.Fatalf("jwt parser should not be called when no auth header is present")
			return nil, nil
		}},
		&authMock.PermissionsCheckerMock{HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
			t.Fatalf("permissions checker should not be called when no auth header is present")
			return false, nil
		}},
		&authMock.ZebedeeClientMock{CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
			t.Fatalf("zebedee client should not be called when no auth header is present")
			return nil, nil
		}},
	)

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetFileMetadataAuth_InvalidJWTReturns401(t *testing.T) {
	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, nil
	}

	handler := api.HandleGetFileMetadataWithPermissions(
		getMetadata,
		&authMock.MiddlewareMock{
			ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
				return nil, errors.New("invalid jwt")
			},
		},
		&authMock.PermissionsCheckerMock{
			HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
				t.Fatalf("permissions checker should not be called when token is invalid")
				return false, nil
			},
		},
		&authMock.ZebedeeClientMock{
			CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
				return nil, errors.New("invalid service token")
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-invalid-jwt-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetFileMetadataAuth_ValidJWTNotAuthorisedReturns403(t *testing.T) {
	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{DatasetID: "dataset-1", Edition: "edition-1"},
		}, nil
	}

	handler := api.HandleGetFileMetadataWithPermissions(
		getMetadata,
		&authMock.MiddlewareMock{ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			return &permissionsAPISDK.EntityData{UserID: "user-1"}, nil
		}},
		&authMock.PermissionsCheckerMock{HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
			return false, nil
		}},
		&authMock.ZebedeeClientMock{CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
			t.Fatalf("zebedee client should not be called for jwt tokens")
			return nil, nil
		}},
	)

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-valid-jwt-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGetFileMetadataAuth_ValidJWTAuthorisedReturns200(t *testing.T) {
	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{DatasetID: "dataset-1", Edition: "edition-1"},
		}, nil
	}

	handler := api.HandleGetFileMetadataWithPermissions(
		getMetadata,
		&authMock.MiddlewareMock{ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			return &permissionsAPISDK.EntityData{UserID: "user-1"}, nil
		}},
		&authMock.PermissionsCheckerMock{HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
			return true, nil
		}},
		&authMock.ZebedeeClientMock{CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
			t.Fatalf("zebedee client should not be called for jwt tokens")
			return nil, nil
		}},
	)

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-valid-jwt-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileMetadataAuth_ServiceTokenAuthorisedReturns200(t *testing.T) {
	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{DatasetID: "dataset-1", Edition: "edition-1"},
		}, nil
	}

	handler := api.HandleGetFileMetadataWithPermissions(
		getMetadata,
		&authMock.MiddlewareMock{
			ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
				return nil, errors.New("not a jwt") // force fallback to service identity
			},
		},
		&authMock.PermissionsCheckerMock{
			HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
				return true, nil
			},
		},
		&authMock.ZebedeeClientMock{
			CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
				return &dprequest.IdentityResponse{Identifier: "service-user"}, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-service")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileMetadataAuth_ServiceTokenNotAuthorisedReturns403(t *testing.T) {
	getMetadata := func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			ContentItem: &files.StoredContentItem{DatasetID: "dataset-1", Edition: "edition-1"},
		}, nil
	}

	handler := api.HandleGetFileMetadataWithPermissions(
		getMetadata,
		&authMock.MiddlewareMock{
			ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
				return nil, errors.New("not a jwt")
			},
		},
		&authMock.PermissionsCheckerMock{
			HasPermissionFunc: func(ctx context.Context, entityData permissionsAPISDK.EntityData, permission string, attributes map[string]string) (bool, error) {
				return false, nil
			},
		},
		&authMock.ZebedeeClientMock{
			CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
				return &dprequest.IdentityResponse{Identifier: "service-user"}, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-service")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
