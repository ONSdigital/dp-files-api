package api_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	permsdk "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetFileMetadataHandlesUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{}, errors.New("broken")
	}, nil, "")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}

func TestGetFileMetadataReturnsUnauthorizedWithoutJwt(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req = mux.SetURLVars(req, map[string]string{"path": "path.jpg"})

	getMetadataCalled := false
	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		getMetadataCalled = true
		return files.StoredRegisteredMetaData{}, nil
	}, &authMock.MiddlewareMock{}, api.PermissionStaticFilesRead)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, getMetadataCalled)
}

func TestGetFileMetadataReturnsUnauthorizedForServiceToken(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req = mux.SetURLVars(req, map[string]string{"path": "path.jpg"})
	req.Header.Set(dprequest.AuthHeaderKey, dprequest.BearerPrefix+"service-token")

	getMetadataCalled := false
	authMiddleware := &authMock.MiddlewareMock{
		ParseFunc: func(tokenString string) (*permsdk.EntityData, error) {
			return &permsdk.EntityData{UserID: "service-1"}, nil
		},
	}
	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		getMetadataCalled = true
		return files.StoredRegisteredMetaData{}, nil
	}, authMiddleware, api.PermissionStaticFilesRead)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, getMetadataCalled)
}

func TestGetFileMetadataIncludesDatasetEditionAttribute(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req = mux.SetURLVars(req, map[string]string{"path": "path.jpg"})
	req.Header.Set(dprequest.AuthHeaderKey, dprequest.BearerPrefix+"test-token")

	authMiddleware := &authMock.MiddlewareMock{
		ParseFunc: func(tokenString string) (*permsdk.EntityData, error) {
			return &permsdk.EntityData{UserID: "user-1", Groups: []string{"groups/role-admin"}}, nil
		},
		RequireWithAttributesFunc: func(permission string, handlerFunc http.HandlerFunc, getAttributes auth.GetAttributesFromRequest) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				attributes, err := getAttributes(r)
				assert.NoError(t, err)
				assert.Equal(t, map[string]string{"dataset_edition": "dataset-1/2025"}, attributes)
				handlerFunc(w, r)
			}
		},
	}

	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			Path: "path.jpg",
			ContentItem: &files.StoredContentItem{
				DatasetID: "dataset-1",
				Edition:   "2025",
			},
		}, nil
	}, authMiddleware, api.PermissionStaticFilesRead)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, authMiddleware.RequireWithAttributesCalls(), 1)
}

func TestGetFileMetadataReturnsForbiddenWhenPermissionDenied(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req = mux.SetURLVars(req, map[string]string{"path": "path.jpg"})
	req.Header.Set(dprequest.AuthHeaderKey, dprequest.BearerPrefix+"test-token")

	authMiddleware := &authMock.MiddlewareMock{
		ParseFunc: func(tokenString string) (*permsdk.EntityData, error) {
			return &permsdk.EntityData{UserID: "user-1", Groups: []string{"groups/role-admin"}}, nil
		},
		RequireWithAttributesFunc: func(permission string, handlerFunc http.HandlerFunc, getAttributes auth.GetAttributesFromRequest) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}
		},
	}

	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		return files.StoredRegisteredMetaData{
			Path: "path.jpg",
			ContentItem: &files.StoredContentItem{
				DatasetID: "dataset-1",
				Edition:   "2025",
			},
		}, nil
	}, authMiddleware, api.PermissionStaticFilesRead)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGetFileMetadataReturnsUnauthorizedForInvalidJwt(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/path.jpg", http.NoBody)
	req = mux.SetURLVars(req, map[string]string{"path": "path.jpg"})
	req.Header.Set(dprequest.AuthHeaderKey, dprequest.BearerPrefix+"invalid-token")

	getMetadataCalled := false
	authMiddleware := &authMock.MiddlewareMock{
		ParseFunc: func(tokenString string) (*permsdk.EntityData, error) {
			return nil, errors.New("invalid token")
		},
	}
	h := api.HandleGetFileMetadata(func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
		getMetadataCalled = true
		return files.StoredRegisteredMetaData{}, nil
	}, authMiddleware, api.PermissionStaticFilesRead)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, getMetadataCalled)
}
