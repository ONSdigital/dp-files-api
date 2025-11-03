package files

import "time"

// FileEvent represents a file access event for the audit log
type FileEvent struct {
	CreatedAt   *time.Time    `json:"created_at,omitempty" bson:"created_at,omitempty"`
	RequestedBy *RequestedBy  `json:"requested_by" bson:"requested_by"`
	Action      string        `json:"action" bson:"action"`
	Resource    string        `json:"resource" bson:"resource"`
	File        *FileMetaData `json:"file" bson:"file"`
}

// RequestedBy represents the user who made the request
type RequestedBy struct {
	ID    string `json:"id" bson:"id"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
}

// FileMetaData represents the basic file information stored in an event
type FileMetaData struct {
	Path          string  `json:"path" bson:"path"`
	IsPublishable bool    `json:"is_publishable" bson:"is_publishable"`
	CollectionID  *string `json:"collection_id,omitempty" bson:"collection_id,omitempty"`
	BundleID      *string `json:"bundle_id,omitempty" bson:"bundle_id,omitempty"`
	Title         string  `json:"title" bson:"title"`
	SizeInBytes   uint64  `json:"size_in_bytes" bson:"size_in_bytes"`
	Type          string  `json:"type" bson:"type"`
	Licence       string  `json:"licence" bson:"licence"`
	LicenceURL    string  `json:"licence_url" bson:"licence_url"`
	State         string  `json:"state,omitempty" bson:"state,omitempty"`
	Etag          string  `json:"etag,omitempty" bson:"etag,omitempty"`
}

// EventsList represents a paginated list of file events
type EventsList struct {
	Count      int         `json:"count"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	TotalCount int         `json:"total_count"`
	Items      []FileEvent `json:"items"`
}

// Action consts
const (
	ActionCreate = "CREATE"
	ActionRead   = "READ"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)
