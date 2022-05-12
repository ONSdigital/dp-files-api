package store

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-files-api/files"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

func (store *Store) GetFileMetadata(ctx context.Context, path string) (files.StoredRegisteredMetaData, error) {
	metadata := files.StoredRegisteredMetaData{}

	err := store.mongoCollection.FindOne(ctx, bson.M{fieldPath: path}, &metadata)
	if err != nil && errors.Is(err, mongodriver.ErrNoDocumentFound) {
		log.Error(ctx, "file metadata not found", err, log.Data{"path": path})
		return metadata, ErrFileNotRegistered
	}

	return metadata, err
}

func (store *Store) GetFilesMetadata(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error) {
	files := make([]files.StoredRegisteredMetaData, 0)
	_, err := store.mongoCollection.Find(ctx, bson.M{fieldCollectionID: collectionID}, &files)

	return files, err
}
