package sdk

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestCreateFileEvent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/file-events", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123","requested_by":{"id":"user123"},"action":"READ","resource":"/downloads/file.csv","file":{"path":"file.csv","type":"csv"}}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.NoError(t, err)
	assert.NotNil(t, createdEvent)
	assert.Equal(t, "user123", createdEvent.RequestedBy.ID)
	assert.Equal(t, files.ActionRead, createdEvent.Action)
	assert.Equal(t, "/downloads/file.csv", createdEvent.Resource)
	assert.Equal(t, "file.csv", createdEvent.File.Path)
}

func TestCreateFileEvent_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code":"ValidationError","description":"invalid event"}]}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestCreateFileEvent_Unauthorised(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":[{"code":"Unauthorised","description":"invalid token"}]}`))
	}))
	defer server.Close()

	client := New(server.URL, "invalid-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "unauthorised")
}

func TestCreateFileEvent_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors":[{"code":"Forbidden","description":"access denied"}]}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCreateFileEvent_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code":"InternalError","description":"database error"}]}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "InternalError")
}

func TestCreateFileEvent_NetworkError(t *testing.T) {
	client := New("http://localhost:99999", "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "failed to execute request")
}

func TestCreateFileEvent_VerifyRequestBody(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"requested_by":{"id":"user123","email":"user@example.com"},"action":"READ","resource":"/downloads/file.csv","file":{"path":"file.csv","type":"text/csv"}}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123", Email: "user@example.com"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "text/csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.NoError(t, err)
	assert.NotNil(t, createdEvent)
	assert.Contains(t, string(capturedBody), "user123")
	assert.Contains(t, string(capturedBody), "user@example.com")
	assert.Contains(t, string(capturedBody), "READ")
	assert.Contains(t, string(capturedBody), "/downloads/file.csv")
}

func TestCreateFileEvent_EmptyErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[]}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), `{"errors":[]}`)
}

func TestCreateFileEvent_EmptyBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "API returned status 500 with no error message")
}

func TestCreateFileEvent_PlainTextError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")

	event := files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	createdEvent, err := client.CreateFileEvent(context.Background(), event)

	assert.Error(t, err)
	assert.Nil(t, createdEvent)
	assert.Contains(t, err.Error(), "Internal Server Error")
}
