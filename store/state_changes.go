package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	StateCreated   = "CREATED"
	StateUploaded  = "UPLOADED"
	StatePublished = "PUBLISHED"
	StateDecrypted = "DECRYPTED"
)

func (store *Store) RegisterFileUpload(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
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

func (store *Store) MarkUploadComplete(ctx context.Context, metaData files.FileEtagChange) error {
	return store.updateStatus(ctx, metaData.Path, metaData.Etag, StateUploaded, StateCreated, "upload_completed_at")
}

func (store *Store) MarkFileDecrypted(ctx context.Context, metaData files.FileEtagChange) error {
	return store.updateStatus(ctx, metaData.Path, metaData.Etag, StateDecrypted, StatePublished, "decrypted_at")
}

func (store *Store) MarkFilePublished(ctx context.Context, path string) error {
	m := files.StoredRegisteredMetaData{}
	err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &m)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "mark file as published: attempted to operate on unregistered file", err, log.Data{"path": path})
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to mark file as published", err, log.Data{"path": path})
		return err
	}

	if m.CollectionID == nil {
		err := ErrCollectionIDNotSet
		log.Error(ctx, "file had no collection id", err, log.Data{"metadata": m})
		return err
	}

	if m.State != StateUploaded {
		log.Error(ctx, fmt.Sprintf("mark file published: file was not in state %s", StateUploaded),
			ErrFileNotInUploadedState, log.Data{"path": path, "current_state": m.State})
		return ErrFileNotInUploadedState
	}

	if m.IsPublishable != true {
		log.Error(ctx, "mark file published: file not set as publishable",
			ErrFileIsNotPublishable, log.Data{"path": path, "is_publishable": m.IsPublishable})
		return ErrFileIsNotPublishable
	}

	_, err = store.mongoCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"state", StatePublished},
				{"last_modified", store.clock.GetCurrentTime()},
				{"published_at", store.clock.GetCurrentTime()}}},
		})

	if err != nil {
		return err
	}

	err = store.kafka.Send(files.AvroSchema, &files.FilePublished{
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
	metadata := files.StoredRegisteredMetaData{}
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

	_, err = store.mongoCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"etag", etag},
				{"state", toState},
				{"last_modified", store.clock.GetCurrentTime()},
				{timestampField, store.clock.GetCurrentTime()}}},
		})

	return err
}
