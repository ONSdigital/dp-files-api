package files

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrDuplicateFile = errors.New("duplicate file path")
var ErrFileNotRegistered = errors.New("file not registered")
var ErrFileNotInCreatedState = errors.New("file state is not in state created")

const (
	stateCreated  = "CREATED"
	stateUploaded = "UPLOADED"
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
		return metadata, ErrFileNotRegistered
	}

	return metadata, err
}

func (s *Store) RegisterFileUpload(ctx context.Context, metaData StoredRegisteredMetaData) error {
	count, err := s.m.Collection(config.MetadataCollection).Count(ctx, bson.M{"path": metaData.Path})
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrDuplicateFile
	}

	metaData.CreatedAt = s.c.GetCurrentTime()
	metaData.LastModified = s.c.GetCurrentTime()
	metaData.State = stateCreated

	_, err = s.m.Collection(config.MetadataCollection).Insert(ctx, metaData)

	return err
}

func (s *Store) MarkUploadComplete(ctx context.Context, metaData StoredUploadCompleteMetaData) error {
	metadata := StoredRegisteredMetaData{}
	err := s.m.Collection(config.MetadataCollection).FindOne(ctx, bson.M{"path": metaData.Path}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return ErrFileNotRegistered
		}
		return err
	}

	if metadata.State != stateCreated {
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

	return err
}
