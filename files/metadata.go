package files

import (
	"context"
	"time"
)

type StoredRegisteredMetaData struct {
	Path              string    `bson:"path" json:"path"`
	IsPublishable     bool      `bson:"is_publishable" json:"is_publishable"`
	CollectionID      string    `bson:"collection_id" json:"collection_id"`
	Title             string    `bson:"title" json:"title"`
	SizeInBytes       uint64    `bson:"size_in_bytes" json:"size_in_bytes"`
	Type              string    `bson:"type" json:"type"`
	Licence           string    `bson:"licence" json:"licence"`
	LicenceUrl        string    `bson:"licence_url" json:"licence_url"`
	CreatedAt         time.Time `bson:"created_at" json:"-"`
	LastModified      time.Time `bson:"last_modified" json:"-"`
	UploadCompletedAt time.Time `bson:"upload_completed_at" json:"-"`
	State             string    `bson:"state" json:"state"`
	Etag              string    `bson:"etag" json:"etag"`
}

type StoredUploadCompleteMetaData struct {
	Path string `bson:"path"`
	Etag string `bson:"etag"`
}

type RegisterFileUpload func(ctx context.Context, metaData StoredRegisteredMetaData) error
type MarkUploadComplete func(ctx context.Context, metaData StoredUploadCompleteMetaData) error
type GetFileMetadata func(ctx context.Context, path string) (StoredRegisteredMetaData, error)
type PublishCollection func(ctx context.Context, collectionID string) error
