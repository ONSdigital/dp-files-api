package files

import (
	"context"
	"time"
)

type StoredMetaData struct {
	Path          string    `bson:"path"`
	IsPublishable bool      `bson:"is_publishable"`
	CollectionID  string    `bson:"collection_id"`
	Title         string    `bson:"title"`
	SizeInBytes   int64     `bson:"size_in_bytes"`
	Type          string    `bson:"type"`
	Licence       string    `bson:"licence"`
	LicenceUrl    string    `bson:"licence_url"`
	createdAt     time.Time `bson:"created_at"`
	lastModified  time.Time `bson:"last_modified"`
	State         string    `bson:"state"`
}

type CreateUploadStartedEntry func(ctx context.Context, metaData StoredMetaData) error
type MakeUploadComplete func(path string) error
