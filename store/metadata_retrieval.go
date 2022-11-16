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

	err := store.metadataCollection.FindOne(ctx, bson.M{fieldPath: path}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "file metadata not found", err, log.Data{"path": path})
			return metadata, ErrFileNotRegistered
		}
		return metadata, err
	}

	// pre-check to avoid fetching collection if it's not necessary
	if metadata.CollectionID == nil || metadata.State != StateUploaded {
		return metadata, nil
	}

	collection, err := store.GetCollection(ctx, *metadata.CollectionID)
	if err != nil {
		return metadata, err
	}

	store.PatchMetadataWithCollectionInfo(&metadata, &collection)

	return metadata, nil
}

func (store *Store) GetFilesMetadata(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error) {
	files := make([]files.StoredRegisteredMetaData, 0)
	_, err := store.metadataCollection.Find(ctx, bson.M{fieldCollectionID: collectionID}, &files)
	if err != nil {
		return nil, err
	}

	collection, err := store.GetCollection(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(files); i++ {
		store.PatchMetadataWithCollectionInfo(&files[i], &collection)
	}

	return files, nil
}

func (store *Store) GetCollection(ctx context.Context, id string) (files.StoredCollection, error) {
	collection := files.StoredCollection{}
	err := store.collectionsCollection.FindOne(ctx, bson.M{fieldID: id}, &collection)
	if err != nil {
		log.Error(ctx, "collection metadata not found", err, log.Data{"id": id})
	}
	return collection, err
}

// For the optimisation purposes, we store the Florence collection publishing information in a separate DB collection.
// Because of this, we need to patch the file metadata in a specific case documented below.
func (store *Store) PatchMetadataWithCollectionInfo(metadata *files.StoredRegisteredMetaData, collection *files.StoredCollection) {
	if metadata == nil || collection == nil {
		return
	}
	// sanity check - collection data should only apply if the collection of the file matches
	// the collection passed in the parameter
	if metadata.CollectionID == nil || *metadata.CollectionID != collection.ID {
		return
	}

	// collection state should only affect the file metadata if the file is in uploaded state
	if metadata.State != StateUploaded {
		return
	}
	// collection state should only affect the file metadata if the collection is in published state
	if collection.State != StatePublished {
		return
	}

	// We now know the file is uploaded and the collection is published. This means the file
	// should be considered published.
	// Also, collection publishing always happens after uploading the file and so the publishing
	// and modification date of the file should be adjusted to match that of the collection.
	metadata.State = StatePublished
	metadata.PublishedAt = collection.PublishedAt
	metadata.LastModified = collection.LastModified
}
