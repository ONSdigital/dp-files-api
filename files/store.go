package files

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrDuplicateFile           = errors.New("duplicate file path")
	ErrFileNotRegistered       = errors.New("file not registered")
	ErrFileNotInCreatedState   = errors.New("file state is not in state created")
	ErrFileNotInUploadedState  = errors.New("file state is not in state uploaded")
	ErrFileNotInPublishedState = errors.New("file state is not in state published")
	ErrNoFilesInCollection     = errors.New("no files found in collection")
	ErrCollectionIDAlreadySet  = errors.New("collection ID already set")
	ErrCollectionIDNotSet      = errors.New("collection ID not set")
)

const (
	StateCreated   = "CREATED"
	StateUploaded  = "UPLOADED"
	StatePublished = "PUBLISHED"
	StateDecrypted = "DECRYPTED"
)

type Store struct {
	mongoCollection MongoCollection
	kafka           kafka.IProducer
	clock           clock.Clock
}

func NewStore(collection MongoCollection, kafkaProducer kafka.IProducer, clk clock.Clock) *Store {
	return &Store{collection, kafkaProducer, clk}
}

func (store *Store) GetFileMetadata(ctx context.Context, path string) (StoredRegisteredMetaData, error) {
	metadata := StoredRegisteredMetaData{}

	err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &metadata)
	if err != nil && errors.Is(err, mongodriver.ErrNoDocumentFound) {
		log.Error(ctx, "file metadata not found", err, log.Data{"path": path})
		return metadata, ErrFileNotRegistered
	}

	return metadata, err
}

func (store *Store) GetFilesMetadata(ctx context.Context, collectionID string) ([]StoredRegisteredMetaData, error) {
	files := []StoredRegisteredMetaData{}
	_, err := store.mongoCollection.Find(ctx, bson.M{"collection_id": collectionID}, &files)

	return files, err
}

func (store *Store) RegisterFileUpload(ctx context.Context, metaData StoredRegisteredMetaData) error {
	count, err := store.mongoCollection.Count(ctx, bson.M{"path": metaData.Path})
	if err != nil {
		log.Error(ctx, "mongo driver count error", err, log.Data{"path": metaData.Path})
		return err
	}

	if count > 0 {
		log.Error(ctx, "file upload already registered", err, log.Data{"path": metaData.Path})
		return ErrDuplicateFile
	}

	metaData.CreatedAt = store.clock.GetCurrentTime()
	metaData.LastModified = store.clock.GetCurrentTime()
	metaData.State = StateCreated

	_, err = store.mongoCollection.Insert(ctx, metaData)
	if err != nil {
		log.Error(ctx, "failed to insert metadata", err, log.Data{"collection": config.MetadataCollection, "metadata": metaData})
		return err
	}

	log.Info(ctx, "registering new file upload", log.Data{"path": metaData.Path})
	return nil
}

func (store *Store) MarkUploadComplete(ctx context.Context, metaData FileEtagChange) error {
	return store.updateStatus(ctx, metaData.Path, metaData.Etag, StateUploaded, StateCreated, "upload_completed_at")
}

func (store *Store) MarkCollectionPublished(ctx context.Context, collectionID string) error {
	count, err := store.mongoCollection.Count(ctx, bson.M{"collection_id": collectionID})
	if err != nil {
		log.Error(ctx, "failed to count files collection", err, log.Data{"collection_id": collectionID})
		return err
	}

	if count == 0 {
		log.Info(ctx, "no files found in collection", log.Data{"collection_id": collectionID})
		return ErrNoFilesInCollection
	}

	count, err = store.mongoCollection.Count(ctx, createCollectionContainsNotUploadedFilesQuery(collectionID))

	if err != nil {
		log.Error(ctx, "failed to count unpublishable files", err, log.Data{"collection_id": collectionID})
		return err
	}

	if count > 0 {
		event := fmt.Sprintf("can not publish collection, not all files in %s state", StateUploaded)
		log.Info(ctx, event, log.Data{"collection_id": collectionID, "num_file_not_state_uploaded": count})
		return ErrFileNotInUploadedState
	}

	_, err = store.mongoCollection.UpdateMany(
		ctx,
		bson.M{"collection_id": collectionID},
		bson.D{
			{"$set", bson.D{
				{"state", StatePublished},
				{"last_modified", store.clock.GetCurrentTime()},
				{"published_at", store.clock.GetCurrentTime()}}},
		})

	if err != nil {
		event := fmt.Sprintf("failed to change files to %s state", StatePublished)
		log.Error(ctx, event, err, log.Data{"collection_id": collectionID})
		return err
	}

	col := make([]StoredRegisteredMetaData, 0)
	store.mongoCollection.Find(ctx, bson.M{"collection_id": collectionID}, &col)
	for _, m := range col {
		err = store.kafka.Send(avroSchema, &FilePublished{
			Path:        m.Path,
			Etag:        m.Etag,
			Type:        m.Type,
			SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
		})

		if err != nil {
			logdata := log.Data{}
			logdata["metadata"] = m
			log.Error(ctx, "sending published message to kafka", err, logdata)
		}
	}

	return nil
}

func (store *Store) MarkFileDecrypted(ctx context.Context, metaData FileEtagChange) error {
	return store.updateStatus(ctx, metaData.Path, metaData.Etag, StateDecrypted, StatePublished, "decrypted_at")
}

func (store *Store) UpdateCollectionID(ctx context.Context, path, collectionID string) error {
	metadata := StoredRegisteredMetaData{}
	err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "update collection ID: attempted to operate on unregistered file", err, log.Data{"path": path})
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to update collection ID", err, log.Data{"path": path})
		return err
	}

	if metadata.CollectionID != nil {
		err := ErrCollectionIDAlreadySet
		log.Error(
			ctx, "update collection ID: collection ID already set",
			err, log.Data{"path": path, "collection_id": *metadata.CollectionID},
		)
		return err
	}

	store.mongoCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"collection_id", collectionID}},
			},
		})

	return nil
}

func (store *Store) MarkFilePublished(ctx context.Context, path string) error {
	m := StoredRegisteredMetaData{}
	err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &m)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "mark file as published: attempted to operate on unregistered file", err, log.Data{"path": path})
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding m to mark file as published", err, log.Data{"path": path})
		return err
	}

	if m.CollectionID == nil {
		err := ErrCollectionIDNotSet
		log.Error(ctx, "file had no collection id", err, log.Data{"metadata": m})
		return err
	}

	if m.State != StateUploaded {
		log.Error(ctx, fmt.Sprintf("mark file published: file was not in state %s", StateUploaded),
			err, log.Data{"path": path, "current_state": m.State})
		return ErrFileNotInUploadedState
	}
	store.mongoCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"state", StatePublished},
				{"last_modified", store.clock.GetCurrentTime()},
				{"published_at", store.clock.GetCurrentTime()}}},
		})
	err = store.kafka.Send(avroSchema, &FilePublished{
		Path:        m.Path,
		Etag:        m.Etag,
		Type:        m.Type,
		SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
	})

	if err != nil {
		return err
	}

	return nil
}

func (store *Store) updateStatus(ctx context.Context, path, etag, toState, expectedCurrentState, timestampField string) error {
	metadata := StoredRegisteredMetaData{}
	err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "mark file as decrypted: attempted to operate on unregistered file", err, log.Data{"path": path})
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to mark file as decrypted", err, log.Data{"path": path})
		return err
	}

	if metadata.State != expectedCurrentState {
		log.Error(ctx, fmt.Sprintf("mark file decrypted: file was not in state %s", StateCreated),
			err, log.Data{"path": path, "current_state": metadata.State})
		return ErrFileNotInPublishedState
	}

	store.mongoCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"etag", etag},
				{"state", toState},
				{"last_modified", store.clock.GetCurrentTime()},
				{timestampField, store.clock.GetCurrentTime()}}},
		})

	return nil
}

func createCollectionContainsNotUploadedFilesQuery(collectionID string) bson.M {
	return bson.M{"$and": []bson.M{
		{"collection_id": collectionID},
		{"state": bson.M{"$ne": StateUploaded}},
	}}
}
