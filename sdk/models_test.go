package sdk_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ONSdigital/dp-files-api/sdk"
	"github.com/stretchr/testify/assert"
)

func TestFileEventMarshalling(t *testing.T) {
	now := time.Now()
	collectionID := "test-collection"

	event := sdk.FileEvent{
		CreatedAt: &now,
		RequestedBy: &sdk.RequestedBy{
			ID:    "user123",
			Email: "user@example.com",
		},
		Action:   sdk.ActionRead,
		Resource: "/downloads/file.csv",
		File: &sdk.FileMetaData{
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

	var event sdk.FileEvent
	err := json.Unmarshal([]byte(jsonStr), &event)

	assert.NoError(t, err)
	assert.Equal(t, "user456", event.RequestedBy.ID)
	assert.Equal(t, "test@example.com", event.RequestedBy.Email)
	assert.Equal(t, sdk.ActionCreate, event.Action)
	assert.Equal(t, "/files/test.xls", event.Resource)
	assert.Equal(t, "test.xls", event.File.Path)
	assert.Equal(t, uint64(2048), event.File.SizeInBytes)
}

func TestActionConstants(t *testing.T) {
	assert.Equal(t, "CREATE", sdk.ActionCreate)
	assert.Equal(t, "READ", sdk.ActionRead)
	assert.Equal(t, "UPDATE", sdk.ActionUpdate)
	assert.Equal(t, "DELETE", sdk.ActionDelete)
}
