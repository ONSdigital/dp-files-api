package api_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/store"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRemoveFileUnsuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", nil)

	errFunc := func(ctx context.Context, path string) error {
		return errors.New("it's all gone very wrong")
	}

	h := api.HandleRemoveFile(errFunc)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	response, _ := io.ReadAll(rec.Body)
	if !bytes.Contains(response, []byte("it's all gone very wrong")) {
		t.Errorf("expected response to contain 'it's all gone very wrong', got %s", response)
	}
}

func TestHandleRemoveFileFileNotRegistered(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", nil)

	errFunc := func(ctx context.Context, path string) error {
		return store.ErrFileNotRegistered
	}

	h := api.HandleRemoveFile(errFunc)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}
	response, _ := io.ReadAll(rec.Body)

	if !bytes.Contains(response, []byte(store.ErrFileNotRegistered.Error())) {
		t.Errorf("expected response to contain '%s', got %s", store.ErrFileNotRegistered.Error(), response)
	}
}

func TestHandleRemoveFileFileIsPublished(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", nil)

	errFunc := func(ctx context.Context, path string) error {
		return store.ErrFileIsPublished
	}

	h := api.HandleRemoveFile(errFunc)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status code %d, got %d", http.StatusConflict, rec.Code)
	}
	response, _ := io.ReadAll(rec.Body)

	if !bytes.Contains(response, []byte(store.ErrFileIsPublished.Error())) {
		t.Errorf("expected response to contain '%s', got %s", store.ErrFileIsPublished.Error(), response)
	}
}

func TestHandleRemoveFileSuccessful(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/path.txt", nil)

	errFunc := func(ctx context.Context, path string) error {
		return nil
	}

	h := api.HandleRemoveFile(errFunc)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty response body, got %s", rec.Body.String())
	}
}
