package sdk_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-files-api/sdk"
	"github.com/stretchr/testify/assert"
)

func TestCreateFileEvent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/file-events", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	client := sdk.New(server.URL, "test-token")

	event := sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := client.CreateFileEvent(context.Background(), event)

	assert.NoError(t, err)
}

func TestCreateFileEvent_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code":"ValidationError","description":"invalid event"}]}`))
	}))
	defer server.Close()

	client := sdk.New(server.URL, "test-token")

	event := sdk.FileEvent{}

	err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestCreateFileEvent_Unauthorised(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":[{"code":"Unauthorised","description":"invalid token"}]}`))
	}))
	defer server.Close()

	client := sdk.New(server.URL, "invalid-token")

	event := sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorised")
}

func TestCreateFileEvent_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors":[{"code":"Forbidden","description":"access denied"}]}`))
	}))
	defer server.Close()

	client := sdk.New(server.URL, "test-token")

	event := sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCreateFileEvent_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code":"InternalError","description":"database error"}]}`))
	}))
	defer server.Close()

	client := sdk.New(server.URL, "test-token")

	event := sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "InternalError")
}

func TestCreateFileEvent_NetworkError(t *testing.T) {
	client := sdk.New("http://localhost:99999", "test-token")

	event := sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute request")
}

func TestCreateFileEvent_VerifyRequestBody(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := sdk.New(server.URL, "test-token")

	event := sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123", Email: "user@example.com"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "text/csv"},
	}

	err := client.CreateFileEvent(context.Background(), event)

	assert.NoError(t, err)
	assert.Contains(t, string(capturedBody), "user123")
	assert.Contains(t, string(capturedBody), "user@example.com")
	assert.Contains(t, string(capturedBody), "READ")
	assert.Contains(t, string(capturedBody), "/downloads/file.csv")
}
