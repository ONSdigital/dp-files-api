package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/stretchr/testify/assert"
)

func TestGetFileEventsSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events", http.NoBody)

	mockEventsList := &files.EventsList{
		Count:      1,
		Limit:      20,
		Offset:     0,
		TotalCount: 1,
		Items:      []files.FileEvent{},
	}

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return mockEventsList, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestGetFileEventsWithLimitAndOffset(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?limit=10&offset=5", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		assert.Equal(t, 10, limit)
		assert.Equal(t, 5, offset)
		return &files.EventsList{Count: 0, Limit: 10, Offset: 5, TotalCount: 0, Items: []files.FileEvent{}}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileEventsWithPath(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?path=test-file.csv", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		assert.Equal(t, "test-file.csv", path)
		return &files.EventsList{Count: 0, Limit: 20, Offset: 0, TotalCount: 0, Items: []files.FileEvent{}}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileEventsWithAfterAndBefore(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?after=2025-01-01T00:00:00Z&before=2025-12-31T23:59:59Z", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		assert.NotNil(t, after)
		assert.NotNil(t, before)
		assert.Equal(t, 2025, after.Year())
		assert.Equal(t, 2025, before.Year())
		return &files.EventsList{Count: 0, Limit: 20, Offset: 0, TotalCount: 0, Items: []files.FileEvent{}}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileEventsWithInvalidLimit(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?limit=abc", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithLimitTooLarge(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?limit=2000", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithNegativeLimit(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?limit=-5", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithInvalidOffset(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?offset=xyz", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithNegativeOffset(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?offset=-10", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithInvalidAfterDatetime(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?after=not-a-date", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithInvalidBeforeDatetime(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?before=invalid-datetime", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetFileEventsWithPathNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?path=nonexistent-file.csv", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, store.ErrPathNotFound
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetFileEventsWithStoreError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		return nil, errors.New("database error")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetFileEventsDefaultPagination(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		assert.Equal(t, 20, limit)
		assert.Equal(t, 0, offset)
		assert.Nil(t, after)
		assert.Nil(t, before)
		assert.Equal(t, "", path)
		return &files.EventsList{Count: 0, Limit: 20, Offset: 0, TotalCount: 0, Items: []files.FileEvent{}}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFileEventsWithAllQueryParams(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file-events?limit=50&offset=10&path=data.csv&after=2025-01-01T00:00:00Z&before=2025-12-31T23:59:59Z", http.NoBody)

	h := api.HandlerGetFileEvents(func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error) {
		assert.Equal(t, 50, limit)
		assert.Equal(t, 10, offset)
		assert.Equal(t, "data.csv", path)
		assert.NotNil(t, after)
		assert.NotNil(t, before)
		return &files.EventsList{Count: 0, Limit: 50, Offset: 10, TotalCount: 0, Items: []files.FileEvent{}}, nil
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
