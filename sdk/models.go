package sdk

import "time"

// FileEvent represents a file access event for the audit log
type FileEvent struct {
	CreatedAt   *time.Time    `json:"created_at,omitempty"`
	RequestedBy *RequestedBy  `json:"requested_by"`
	Action      string        `json:"action"`
	Resource    string        `json:"resource"`
	File        *FileMetaData `json:"file"`
}

// RequestedBy represents the user who made the request
type RequestedBy struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

// FileMetaData represents the basic file information stored in an event
type FileMetaData struct {
	Path          string  `json:"path"`
	IsPublishable bool    `json:"is_publishable"`
	CollectionID  *string `json:"collection_id,omitempty"`
	BundleID      *string `json:"bundle_id,omitempty"`
	Title         string  `json:"title"`
	SizeInBytes   uint64  `json:"size_in_bytes"`
	Type          string  `json:"type"`
	Licence       string  `json:"licence"`
	LicenceURL    string  `json:"licence_url"`
	State         string  `json:"state,omitempty"`
	Etag          string  `json:"etag,omitempty"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// Error represents a single error in the error response
type Error struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Action const
const (
	ActionCreate = "CREATE"
	ActionRead   = "READ"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)
