package files

import (
	"time"
)

type StoredRegisteredMetaData struct {
	Path              string     `bson:"path" json:"path"`
	IsPublishable     bool       `bson:"is_publishable" json:"is_publishable"`
	CollectionID      *string    `bson:"collection_id,omitempty" json:"collection_id,omitempty"`
	Title             string     `bson:"title" json:"title"`
	SizeInBytes       uint64     `bson:"size_in_bytes" json:"size_in_bytes"`
	Type              string     `bson:"type" json:"type"`
	Licence           string     `bson:"licence" json:"licence"`
	LicenceUrl        string     `bson:"licence_url" json:"licence_url"`
	CreatedAt         time.Time  `bson:"created_at" json:"-"`
	LastModified      time.Time  `bson:"last_modified" json:"-"`
	UploadCompletedAt *time.Time `bson:"upload_completed_at,omitempty" json:"-"`
	PublishedAt       *time.Time `bson:"published_at,omitempty" json:"-"`
	DecryptedAt       *time.Time `bson:"decrypted_at,omitempty" json:"-"`
	State             string     `bson:"state" json:"state"`
	Etag              string     `bson:"etag" json:"etag"`
}

type StoredCollection struct {
	ID           string     `bson:"id" json:"id"`
	State        string     `bson:"state" json:"state"`
	LastModified time.Time  `bson:"last_modified" json:"-"`
	PublishedAt  *time.Time `bson:"published_at,omitempty" json:"-"`
}

type FileEtagChange struct {
	Path string
	Etag string
}
