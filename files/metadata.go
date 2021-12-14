package files

import (
	"context"
	"time"
)

type MetaData struct {
	Path          string    `json:"path" bson:"path"`
	IsPublishable bool      `json:"is_publishable" bson:"is_publishable"`
	CollectionID  string    `json:"collection_id" bson:"collection_id"`
	Title         string    `json:"title" bson:"title"`
	SizeInBytes   int64     `json:"size_in_bytes" bson:"size_in_bytes"`
	Type          string    `json:"type" bson:"type"`
	Licence       string    `json:"licence" bson:"licence"`
	LicenceUrl    string    `json:"licence_url" bson:"licence_url"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	LastModified  time.Time `json:"last_modified" bson:"last_modified"`
	State         string    `json:"state" bson:"state"`
}

type CreateUploadStartedEntry func(ctx context.Context, metaData MetaData) error
type MakeUploadComplete func(path string) error
