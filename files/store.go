package files

import (
	"context"
	"errors"
	"fmt"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrDuplicateFile          = errors.New("duplicate file path")
	ErrFileNotRegistered      = errors.New("file not registered")
	ErrFileNotInCreatedState  = errors.New("file state is not in state created")
	ErrFileNotInUploadedState = errors.New("file state is not in state uploaded")
	ErrNoFilesInCollection    = errors.New("no files found in collection")
)

const (
	stateCreated   = "CREATED"
	stateUploaded  = "UPLOADED"
	statePublished = "PUBLISHED"
)

type Store struct {
	m mongo.Client
	c clock.Clock
}

func NewStore(m mongo.Client, c clock.Clock) *Store {
	return &Store{m, c}
}

func (s *Store) GetFileMetadata(ctx context.Context, path string) (StoredRegisteredMetaData, error) {
	metadata := StoredRegisteredMetaData{}

	err := s.m.Collection(config.MetadataCollection).FindOne(ctx, bson.M{"path": path}, &metadata)
	if err != nil && errors.Is(err, mongodriver.ErrNoDocumentFound) {
		log.Error(ctx, "file metadata not found", err, log.Data{"path": path})
		return metadata, ErrFileNotRegistered
	}

	return metadata, err
}

func (s *Store) RegisterFileUpload(ctx context.Context, metaData StoredRegisteredMetaData) error {
	count, err := s.m.Collection(config.MetadataCollection).Count(ctx, bson.M{"path": metaData.Path})
	if err != nil {
		log.Error(ctx, "mongo driver count error", err, log.Data{"path": metaData.Path})
		return err
	}

	if count > 0 {
		log.Error(ctx, "file upload already registered", err, log.Data{"path": metaData.Path})
		return ErrDuplicateFile
	}

	metaData.CreatedAt = s.c.GetCurrentTime()
	metaData.LastModified = s.c.GetCurrentTime()
	metaData.State = stateCreated

	_, err = s.m.Collection(config.MetadataCollection).Insert(ctx, metaData)
	if err != nil {
		log.Error(ctx, "failed to insert metadata", err, log.Data{"collection": config.MetadataCollection, "metadata": metaData})
		return err
	}

	log.Info(ctx, "registering new file upload", log.Data{"path": metaData.Path})
	return nil
}

func (s *Store) MarkUploadComplete(ctx context.Context, metaData StoredUploadCompleteMetaData) error {
	metadata := StoredRegisteredMetaData{}
	err := s.m.Collection(config.MetadataCollection).FindOne(ctx, bson.M{"path": metaData.Path}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "mark upload complete: attempted to operate on unregistered file", err, log.Data{"path": metaData.Path})
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to mark upload complete", err, log.Data{"path": metaData.Path})
		return err
	}

	if metadata.State != stateCreated {
		log.Error(ctx, fmt.Sprintf("mark upload complete: file was not in state %s", stateCreated),
			err, log.Data{"path": metaData.Path, "current_state": metadata.State})
		return ErrFileNotInCreatedState
	}

	_, err = s.m.Collection(config.MetadataCollection).Update(
		ctx,
		bson.M{"path": metaData.Path},
		bson.D{
			{"$set", bson.D{
				{"etag", metaData.Etag},
				{"state", stateUploaded},
				{"last_modified", s.c.GetCurrentTime()},
				{"upload_completed_at", s.c.GetCurrentTime()}}},
		})

	if err != nil {
		log.Error(ctx, "failed to mark upload complete", err, log.Data{"metadata": metaData, "collection": config.MetadataCollection})
		return err
	}

	log.Info(ctx, "marking file upload complete", log.Data{"path": metaData.Path})
	return nil
}

func (s *Store) PublishCollection(ctx context.Context, collectionID string) error {

	count, err := s.m.Collection(config.MetadataCollection).Count(ctx, bson.M{"collection_id": collectionID})
	if err != nil {
		log.Error(ctx, "failed to count files collection", err, log.Data{"collection_id": collectionID})
		return err
	}

	if count == 0 {
		log.Info(ctx, "no files found in collection", log.Data{"collection_id": collectionID})
		return ErrNoFilesInCollection
	}

	count, err = s.m.Collection(config.MetadataCollection).
		Count(ctx, createCollectionContainsNotUploadedFilesQuery(collectionID))

	if err != nil {
		log.Error(ctx, "failed to count unpublishable files", err, log.Data{"collection_id": collectionID})
		return err
	}

	if count > 0 {
		event := fmt.Sprintf("can not publish collection, not all files in %s state", stateUploaded)
		log.Info(ctx, event, log.Data{"collection_id": collectionID, "num_file_not_state_uploaded": count})
		return ErrFileNotInUploadedState
	}

	_, err = s.m.Collection(config.MetadataCollection).UpdateMany(
		ctx,
		bson.M{"collection_id": collectionID},
		bson.D{
			{"$set", bson.D{
				{"state", statePublished},
				{"last_modified", s.c.GetCurrentTime()},
				{"published_at", s.c.GetCurrentTime()}}},
		})

	if err != nil {
		event := fmt.Sprintf("failed to change files to %s state", statePublished)
		log.Error(ctx, event, err, log.Data{"collection_id": collectionID})
		return err
	}

	return nil
}

func createCollectionContainsNotUploadedFilesQuery(collectionID string) bson.M {
	return bson.M{"$and": []bson.M{
		{"collection_id": collectionID},
		{"state": bson.M{"$ne": stateUploaded}},
	}}
}
