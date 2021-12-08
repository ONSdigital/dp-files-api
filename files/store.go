package files

import (
	"context"
	"github.com/ONSdigital/dp-files-api/mongo"
)

type Store struct {
	m mongo.Client
}

func NewStore(m mongo.Client) *Store {
	return &Store{m}
}

func (s *Store) CreateUploadStarted(ctx context.Context, metaData MetaData) error {
	_, err := s.m.Connection().C("metadata").Insert(ctx, metaData)

	return err
}




