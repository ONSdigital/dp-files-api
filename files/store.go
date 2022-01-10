package files

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-files-api/clock"
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

func (s *Store) CreateUploadStarted(ctx context.Context, metaData StoredRegisteredMetaData) error {
	count, err := s.m.Connection().C("metadata").Find(bson.M{"path": metaData.Path}).Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrDuplicateFile
	}

	metaData.CreatedAt = s.c.GetCurrentTime()
	metaData.LastModified = s.c.GetCurrentTime()
	metaData.State = stateCreated

	_, err = s.m.Connection().C("metadata").Insert(ctx, metaData)

	return err
}

func (s *Store) MarkUploadComplete(ctx context.Context, metaData StoredUploadCompleteMetaData) error {
	m := StoredRegisteredMetaData{}
	err := s.m.Connection().C("metadata").FindOne(ctx, bson.M{"path": metaData.Path}, &m)
	if err != nil {
		if mongodriver.IsErrNoDocumentFound(err) {
			return ErrFileNotRegistered
		}
		return err
	}

	if m.State != stateCreated {
		return ErrFileNotInCreatedState
	}

	_, err = s.m.Connection().C("metadata").Update(
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
