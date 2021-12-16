package files

import (
	"context"
	"time"
)

type StoredRegisteredMetaData struct {
	Path              string    `bson:"path"`
	IsPublishable     bool      `bson:"is_publishable"`
	CollectionID      string    `bson:"collection_id"`
	Title             string    `bson:"title"`
	SizeInBytes       uint64    `bson:"size_in_bytes"`
	Type              string    `bson:"type"`
	Licence           string    `bson:"licence"`
	LicenceUrl        string    `bson:"licence_url"`
	CreatedAt         time.Time `bson:"created_at"`
	LastModified      time.Time `bson:"last_modified"`
	UploadCompletedAt time.Time `bson:"upload_completed_at"`
	State             string    `bson:"state"`
	Etag              string    `bson:"etag"`
}

type StoredUploadCompleteMetaData struct {
	Path string `bson:"path"`
	Etag string `bson:"etag"`
}

type CreateUploadStartedEntry func(ctx context.Context, metaData StoredRegisteredMetaData) error
type MarkUploadComplete func(ctx context.Context, metaData StoredUploadCompleteMetaData) error
