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
	"github.com/ONSdigital/dp-files-api/mongo"

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
)

const (
	StateCreated  = "CREATED"
	StateUploaded = "UPLOADED"
	StatePublished = "PUBLISHED"
	StateDecrypted = "DECRYPTED"
)

type Store struct {
	m mongo.Client
	k kafka.IProducer
	c clock.Clock
}

func NewStore(m mongo.Client, k kafka.IProducer, c clock.Clock) *Store {
	return &Store{m, k, c}
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
	metaData.State = StateCreated

	_, err = s.m.Collection(config.MetadataCollection).Insert(ctx, metaData)
	if err != nil {
		log.Error(ctx, "failed to insert metadata", err, log.Data{"collection": config.MetadataCollection, "metadata": metaData})
		return err
	}

	log.Info(ctx, "registering new file upload", log.Data{"path": metaData.Path})
	return nil
}

func (s *Store) MarkUploadComplete(ctx context.Context, metaData FileEtagChange) error {
	return s.updateStatus(ctx, metaData.Path, metaData.Etag, StateUploaded, StateCreated, "upload_completed_at")
}

func (s *Store) MarkCollectionPublished(ctx context.Context, collectionID string) error {
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
		event := fmt.Sprintf("can not publish collection, not all files in %s state", StateUploaded)
		log.Info(ctx, event, log.Data{"collection_id": collectionID, "num_file_not_state_uploaded": count})
		return ErrFileNotInUploadedState
	}

	_, err = s.m.Collection(config.MetadataCollection).UpdateMany(
		ctx,
		bson.M{"collection_id": collectionID},
		bson.D{
			{"$set", bson.D{
				{"state", StatePublished},
				{"last_modified", s.c.GetCurrentTime()},
				{"published_at", s.c.GetCurrentTime()}}},
		})

	if err != nil {
		event := fmt.Sprintf("failed to change files to %s state", StatePublished)
		log.Error(ctx, event, err, log.Data{"collection_id": collectionID})
		return err
	}

	col := make([]StoredRegisteredMetaData, 0)
	s.m.Collection(config.MetadataCollection).Find(ctx, bson.M{"collection_id": collectionID}, &col)
	for _, m := range col {
		err = s.k.Send(avroSchema, &FilePublished{
			Path:        m.Path,
			Etag:        m.Etag,
			Type:        m.Type,
			SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
		})

		if err != nil {
			fmt.Printf("PUBLISH ERROR: %s", err.Error())
		} else {
			fmt.Println("MESSAGE SENT SUCCESSFULLY!")
		}
	}

	return nil
}

func (s *Store) MarkFileDecrypted(ctx context.Context, metaData FileEtagChange) error {
	return s.updateStatus(ctx, metaData.Path, metaData.Etag, StateDecrypted, StatePublished, "decrypted_at")
}

func createCollectionContainsNotUploadedFilesQuery(collectionID string) bson.M {
	return bson.M{"$and": []bson.M{
		{"collection_id": collectionID},
		{"state": bson.M{"$ne": StateUploaded}},
	}}
}

func (s *Store) updateStatus(ctx context.Context, path, etag, toState, expectedCurrentState, timestampField string) error {
	metadata := StoredRegisteredMetaData{}
	err := s.m.Collection(config.MetadataCollection).FindOne(ctx, bson.M{"path": path}, &metadata)
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

	s.m.Collection(config.MetadataCollection).Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"etag", etag},
				{"state", toState},
				{"last_modified", s.c.GetCurrentTime()},
				{timestampField, s.c.GetCurrentTime()}}},
		})

	return nil
}
