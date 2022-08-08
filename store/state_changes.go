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
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	StateCreated   = "CREATED"
	StateUploaded  = "UPLOADED"
	StatePublished = "PUBLISHED"
	StateDecrypted = "DECRYPTED"
)

func (store *Store) RegisterFileUpload(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
	logdata := log.Data{"path": metaData.Path}

	metaData.CreatedAt = store.clock.GetCurrentTime()
	metaData.LastModified = store.clock.GetCurrentTime()
	metaData.State = StateCreated

	if _, err := store.mongoCollection.Insert(ctx, metaData); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Error(ctx, "file upload already registered", err, logdata)
			return ErrDuplicateFile
		}
		log.Error(ctx, "failed to insert metadata", err, log.Data{"collection": config.MetadataCollection, "metadata": metaData})
		return err
	}

	log.Info(ctx, "registering new file upload", logdata)
	return nil
}

func (store *Store) MarkUploadComplete(ctx context.Context, metaData files.FileEtagChange) error {
	return store.updateStatus(ctx, metaData.Path, metaData.Etag, StateUploaded, StateCreated, fieldUploadCompletedAt)
}

func (store *Store) MarkFileDecrypted(ctx context.Context, metaData files.FileEtagChange) error {
	return store.updateStatus(ctx, metaData.Path, metaData.Etag, StateDecrypted, StatePublished, fieldDecryptedAt)
}

func (store *Store) MarkFilePublished(ctx context.Context, path string) error {
	m := files.StoredRegisteredMetaData{}
	if err := store.mongoCollection.FindOne(ctx, bson.M{fieldPath: path}, &m); err != nil {
		logdata := log.Data{"path": path}
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "mark file as published: attempted to operate on unregistered file", err, logdata)
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to mark file as published", err, logdata)
		return err
	}

	if m.CollectionID == nil {
		log.Error(ctx, "file had no collection id", ErrCollectionIDNotSet, log.Data{"metadata": m})
		return ErrCollectionIDNotSet
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

	now := store.clock.GetCurrentTime()
	_, err := store.mongoCollection.Update(
		ctx,
		bson.M{fieldPath: path},
		bson.D{
			{"$set", bson.D{
				{fieldState, StatePublished},
				{fieldLastModified, now},
				{fieldPublishedAt, now}}},
		})
	if err != nil {
		return err
	}

	return store.kafka.Send(files.AvroSchema, &files.FilePublished{
		Path:        m.Path,
		Etag:        m.Etag,
		Type:        m.Type,
		SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
	})
}

func (store *Store) updateStatus(ctx context.Context, path, etag, toState, expectedCurrentState, timestampField string) error {
	metadata := files.StoredRegisteredMetaData{}
	if err := store.mongoCollection.FindOne(ctx, bson.M{fieldPath: path}, &metadata); err != nil {
		logdata := log.Data{"path": path}
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "mark file as decrypted: attempted to operate on unregistered file", err, logdata)
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to mark file as decrypted", err, logdata)
		return err
	}

	if metadata.State != expectedCurrentState {
		log.Error(ctx, fmt.Sprintf("mark file decrypted: file was not in state %s", StateCreated),
			ErrFileNotInPublishedState, log.Data{"path": path, "current_state": metadata.State})
		return ErrFileNotInPublishedState
	}

	now := store.clock.GetCurrentTime()
	_, err := store.mongoCollection.Update(
		ctx,
		bson.M{fieldPath: path},
		bson.D{
			{"$set", bson.D{
				{fieldEtag, etag},
				{fieldState, toState},
				{fieldLastModified, now},
				{timestampField, now}}},
		})

	return err
}
