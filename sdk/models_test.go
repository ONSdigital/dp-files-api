package sdk_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/stretchr/testify/assert"
)

func TestFileEventMarshalling(t *testing.T) {
	now := time.Now()
	collectionID := "test-collection"

	event := files.FileEvent{
		CreatedAt: &now,
		RequestedBy: &files.RequestedBy{
			ID:    "user123",
			Email: "user@example.com",
		},
		Action:   files.ActionRead,
		Resource: "/downloads/file.csv",
		File: &files.FileMetaData{
			Path:         "file.csv",
			Type:         "text/csv",
			CollectionID: &collectionID,
			SizeInBytes:  1024,
		},
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "user123")
	assert.Contains(t, string(jsonData), "READ")
}

func TestFileEventUnmarshalling(t *testing.T) {
	jsonStr := `{
		"requested_by": {"id": "user456", "email": "test@example.com"},
		"action": "CREATE",
		"resource": "/files/test.xls",
		"file": {"path": "test.xls", "type": "application/xls", "size_in_bytes": 2048}
	}`

	var event files.FileEvent
	err := json.Unmarshal([]byte(jsonStr), &event)

	assert.NoError(t, err)
	assert.Equal(t, "user456", event.RequestedBy.ID)
	assert.Equal(t, "test@example.com", event.RequestedBy.Email)
	assert.Equal(t, files.ActionCreate, event.Action)
	assert.Equal(t, "/files/test.xls", event.Resource)
	assert.Equal(t, "test.xls", event.File.Path)
	assert.Equal(t, uint64(2048), event.File.SizeInBytes)
}

func TestActionConstants(t *testing.T) {
	assert.Equal(t, "CREATE", files.ActionCreate)
	assert.Equal(t, "READ", files.ActionRead)
	assert.Equal(t, "UPDATE", files.ActionUpdate)
	assert.Equal(t, "DELETE", files.ActionDelete)
}
