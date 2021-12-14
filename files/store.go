package files

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var ErrDuplicateFile = errors.New("duplicate file path")

type Store struct {
	m mongo.Client
}

func NewStore(m mongo.Client) *Store {
	return &Store{m}
}

func (s *Store) CreateUploadStarted(ctx context.Context, metaData MetaData) error {


	finder := s.m.Connection().C("metadata").Find(bson.M{"path": metaData.Path})
	count, err := finder.Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrDuplicateFile
	}

	metaData.CreatedAt = time.Now()
	metaData.LastModified = time.Now()
	metaData.State = "CREATED"

	_, err = s.m.Connection().C("metadata").Insert(ctx, metaData)

	return err
}
