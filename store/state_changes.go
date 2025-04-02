package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	StateCreated   = "CREATED"
	StateUploaded  = "UPLOADED"
	StatePublished = "PUBLISHED"
	StateMoved     = "MOVED"
)

// GetFilesMetadata godoc
// @Description  POSTs metadata for a file when an upload has started.
// @Tags         File upload started
// @Produce      json
// @Param	 	 request formData files.StoredRegisteredMetaData false "StoredRegisteredMetaData"
// @Success      200
// @Failure      400
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /files [post]
func (store *Store) RegisterFileUpload(ctx context.Context, metaData files.StoredRegisteredMetaData) error {
	logdata := log.Data{"path": metaData.Path}

	// don't register file upload if it is already registered
	m := files.StoredRegisteredMetaData{}
	errFindingMetadata := store.metadataCollection.FindOne(ctx, bson.M{fieldPath: metaData.Path}, &m)
	if errFindingMetadata != nil && !errors.Is(errFindingMetadata, mongodriver.ErrNoDocumentFound) {
		log.Error(ctx, "error while finding metadata", errFindingMetadata, logdata)
		return errFindingMetadata
	}

	if m.State == StateUploaded && *m.CollectionID == *metaData.CollectionID {
		log.Info(ctx, "File upload already registered: skipping registration of file metadata", logdata)
		return nil
	}

	// delete existing file metadata if file upload comes from a different collection
	if m.State == StateUploaded && *m.CollectionID != *metaData.CollectionID {
		result, err := store.metadataCollection.Delete(ctx, bson.M{fieldPath: metaData.Path})
		if err != nil {
			log.Error(ctx, "error while deleting metadata", err, logdata)
			return err
		}
		if result.DeletedCount > 0 {
			log.Info(ctx, "deleted existing file metadata", logdata)
		}
	}

	// check to see if collectionID exists and is not-published
	if metaData.CollectionID != nil {
		logdata["collection_id"] = *metaData.CollectionID
		published, err := store.IsCollectionPublished(ctx, *metaData.CollectionID)
		if err != nil {
			log.Error(ctx, "collection published check error", err, logdata)
			return err
		}
		if published {
			log.Error(ctx, "collection is already published", ErrCollectionAlreadyPublished, logdata)
			return ErrCollectionAlreadyPublished
		}
	}
	now := store.clock.GetCurrentTime()
	metaData.CreatedAt = now
	metaData.LastModified = now
	metaData.State = StateCreated

	if _, err := store.metadataCollection.Insert(ctx, metaData); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Error(ctx, "file upload already registered", err, logdata)
			return ErrDuplicateFile
		}
		log.Error(ctx, "failed to insert metadata", err, log.Data{"collection": config.MetadataCollection, "metadata": metaData})
		return err
	}
	if metaData.CollectionID != nil {
		err := store.registerCollection(ctx, *metaData.CollectionID)
		if err != nil {
			log.Error(ctx, "failed to register collection", err, logdata)
			return err
		}
	}

	log.Info(ctx, "registering new file upload", logdata)
	return nil
}

func (store *Store) MarkUploadComplete(ctx context.Context, metaData files.FileEtagChange) error {
	return store.updateFileState(ctx, metaData.Path, metaData.Etag, StateUploaded, StateCreated, fieldUploadCompletedAt)
}

func (store *Store) MarkFileMoved(ctx context.Context, metaData files.FileEtagChange) error {
	return store.updateFileState(ctx, metaData.Path, metaData.Etag, StateMoved, StatePublished, fieldMovedAt)
}

func (store *Store) MarkFilePublished(ctx context.Context, path string) error {
	logdata := log.Data{"path": path}

	m, err := store.GetFileMetadata(ctx, path)
	if err != nil {
		if errors.Is(err, ErrFileNotRegistered) {
			log.Error(ctx, "mark file as published: attempted to operate on unregistered file", err, logdata)
			return ErrFileNotRegistered
		}
		log.Error(ctx, "mark file as published: failed finding file metadata", err, logdata)
		return err
	}
	logdata["metadata"] = m

	if m.State != StateUploaded {
		log.Error(ctx, fmt.Sprintf("mark file published: file was not in state %s", StateUploaded),
			ErrFileNotInUploadedState, logdata)
		return ErrFileNotInUploadedState
	}

	if !m.IsPublishable {
		log.Error(ctx, "mark file published: file not set as publishable",
			ErrFileIsNotPublishable, logdata)
		return ErrFileIsNotPublishable
	}

	now := store.clock.GetCurrentTime()
	_, err = store.metadataCollection.Update(
		ctx,
		bson.M{fieldPath: path},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: fieldState, Value: StatePublished},
				{Key: fieldLastModified, Value: now},
				{Key: fieldPublishedAt, Value: now}}},
		})
	if err != nil {
		return err
	}

	log.Info(ctx, fmt.Sprintf("file set as published - %s", now.String()), logdata)

	return store.kafka.Send(files.AvroSchema, &files.FilePublished{
		Path:        m.Path,
		Etag:        m.Etag,
		Type:        m.Type,
		SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
	})
}

func (store *Store) updateFileState(ctx context.Context, path, etag, toState, expectedCurrentState, timestampField string) error {
	logdata := log.Data{
		"path":                 path,
		"expectedCurrentState": expectedCurrentState,
		"toState":              toState,
	}

	metadata, err := store.GetFileMetadata(ctx, path)
	if err != nil {
		if errors.Is(err, ErrFileNotRegistered) {
			log.Error(ctx, "update file state: attempted to operate on unregistered file", err, logdata)
			return ErrFileNotRegistered
		}
		log.Error(ctx, "update file state: failed finding file metadata", err, logdata)
		return err
	}
	logdata["actualCurrentState"] = metadata.State

	var isCollectionPublished bool
	if metadata.CollectionID != nil {
		isCollectionPublished, err = store.IsCollectionPublished(ctx, *metadata.CollectionID) // also moved
		if err != nil {
			log.Error(ctx, "is collection published: caught db error", err, logdata)
			return err
		}
	}

	// update only timestamps if we are already in uploaded state
	if !isCollectionPublished && metadata.State != StateMoved {
		if toState == StateUploaded && metadata.State == StateUploaded {
			now := store.clock.GetCurrentTime()
			_, err = store.metadataCollection.Update(
				ctx,
				bson.M{fieldPath: path},
				bson.D{
					{Key: "$set", Value: bson.D{
						{Key: fieldEtag, Value: etag},
						{Key: fieldLastModified, Value: now},
						{Key: timestampField, Value: now}}},
				})
			if err != nil {
				log.Error(ctx, "error while updating file metadata", err, logdata)
				return err
			}
			log.Info(ctx, "file metadata updated", logdata)
			return nil
		}
	}

	if metadata.State != expectedCurrentState {
		log.Error(ctx, "update file state: state mismatch", ErrFileStateMismatch, logdata)
		return ErrFileStateMismatch
	}
	// while publishing check that you are publishing the correct/expected version of the file
	if toState == StateMoved {
		head, err := store.s3client.Head(ctx, metadata.Path)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("Failed trying to get head data for %s from bucket %s", metadata.Path, store.cfg.PrivateBucketName), err)
			return err
		}
		if head.ETag != nil && (strings.Trim(*head.ETag, "\"") != metadata.Etag) {
			log.Error(ctx, fmt.Sprintf("Etags mismatch, expected [%s], from s3 [%s]", metadata.Etag, *head.ETag), ErrEtagMismatchWhilePublishing)
			return ErrEtagMismatchWhilePublishing
		}
	}

	now := store.clock.GetCurrentTime()
	_, err = store.metadataCollection.Update(
		ctx,
		bson.M{fieldPath: path},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: fieldEtag, Value: etag},
				{Key: fieldState, Value: toState},
				{Key: fieldLastModified, Value: now},
				{Key: timestampField, Value: now}}},
		})

	return err
}
