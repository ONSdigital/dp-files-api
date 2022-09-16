package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ONSdigital/dp-files-api/files"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

func (store *Store) UpdateCollectionID(ctx context.Context, path, collectionID string) error {
	metadata := files.StoredRegisteredMetaData{}
	logdata := log.Data{"path": path}

	if err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &metadata); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "update collection ID: attempted to operate on unregistered file", err, logdata)
			return ErrFileNotRegistered
		}
		log.Error(ctx, "failed finding metadata to update collection ID", err, logdata)
		return err
	}

	if metadata.CollectionID != nil {
		logdata["collection_id"] = *metadata.CollectionID
		log.Error(ctx, "update collection ID: collection ID already set", ErrCollectionIDAlreadySet, logdata)
		return ErrCollectionIDAlreadySet
	}

	_, err := store.mongoCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{"$set", bson.D{
				{"collection_id", collectionID}},
			},
		})

	return err
}

func (store *Store) MarkCollectionPublished(ctx context.Context, collectionID string) error {
	count, err := store.mongoCollection.Count(ctx, bson.M{fieldCollectionID: collectionID})
	logdata := log.Data{"collection_id": collectionID}
	if err != nil {
		log.Error(ctx, "failed to count files collection", err, logdata)
		return err
	}

	if count == 0 {
		log.Info(ctx, "no files found in collection", logdata)
		return ErrNoFilesInCollection
	}

	count, err = store.mongoCollection.Count(ctx, createCollectionContainsNotUploadedFilesQuery(collectionID))

	if err != nil {
		log.Error(ctx, "failed to count unpublishable files", err, logdata)
		return err
	}

	if count > 0 {
		event := fmt.Sprintf("can not publish collection, not all files in %s state", StateUploaded)
		log.Info(ctx, event, log.Data{"collection_id": collectionID, "num_file_not_state_uploaded": count})
		return ErrFileNotInUploadedState
	}

	now := store.clock.GetCurrentTime()
	_, err = store.mongoCollection.UpdateMany(
		ctx,
		bson.M{fieldCollectionID: collectionID},
		bson.D{
			{"$set", bson.D{
				{fieldState, StatePublished},
				{fieldLastModified, now},
				{fieldPublishedAt, now}}},
		})
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to change files to %s state", StatePublished), err, logdata)
		return err
	}

	requestID := request.GetRequestId(ctx)
	newCtx := request.WithRequestId(context.Background(), requestID)
	go store.NotifyCollectionPublished(newCtx, collectionID)

	return nil
}

func (store *Store) NotifyCollectionPublished(ctx context.Context, collectionID string) {
	log.Info(ctx, "notify collection published start", log.Data{"collection_id": collectionID})

	col := make([]files.StoredRegisteredMetaData, 0)
	if _, err := store.mongoCollection.Find(ctx, bson.M{fieldCollectionID: collectionID}, &col); err != nil {
		log.Error(ctx, "notify collection published: failed to query collection", err, log.Data{"collection_id": collectionID})
		return
	}

	for _, m := range col {
		fp := &files.FilePublished{
			Path:        m.Path,
			Type:        m.Type,
			Etag:        m.Etag,
			SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
		}
		if err := store.kafka.Send(files.AvroSchema, fp); err != nil {
			log.Error(ctx, "notify collection published: can't send message to kafka", err, log.Data{"metadata": m})
		}
	}

	log.Info(ctx, "notify collection published end", log.Data{"collection_id": collectionID})
}
