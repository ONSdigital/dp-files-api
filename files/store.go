package files

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrDuplicateFile = errors.New("duplicate file path")
var ErrFileNotRegistered = errors.New("file not registered")
var ErrFileNotInCreatedState = errors.New("file state is not in state created")

const (
	stateCreated = "CREATED"
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
	finder := s.m.Connection().C("metadata").Find(bson.M{"path": metaData.Path})
	count, err := finder.Count(ctx)
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
	finder := s.m.Connection().C("metadata").Find(bson.M{"path": metaData.Path})
	count, err := finder.Count(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrFileNotRegistered
	}

	m := StoredRegisteredMetaData{}
	err = finder.One(ctx, &m)
	if err != nil {
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
