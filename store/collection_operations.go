package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-files-api/files"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
)

func (store *Store) UpdateCollectionID(ctx context.Context, path, collectionID string) error {
	metadata := files.StoredRegisteredMetaData{}
	err := store.mongoCollection.FindOne(ctx, bson.M{"path": path}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "update collection ID: attempted to operate on unregistered file", err, log.Data{"path": path})
			return ErrFileNotRegistered
		}

		log.Error(ctx, "failed finding metadata to update collection ID", err, log.Data{"path": path})
		return err
	}

	if metadata.CollectionID != nil {
		err := ErrCollectionIDAlreadySet
		log.Error(
			ctx, "update collection ID: collection ID already set",
			err, log.Data{"path": path, "collection_id": *metadata.CollectionID},
		)
		return err
	}

	_, err = store.mongoCollection.Update(
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
	count, err := store.mongoCollection.Count(ctx, bson.M{"collection_id": collectionID})
	if err != nil {
		log.Error(ctx, "failed to count files collection", err, log.Data{"collection_id": collectionID})
		return err
	}

	if count == 0 {
		log.Info(ctx, "no files found in collection", log.Data{"collection_id": collectionID})
		return ErrNoFilesInCollection
	}

	count, err = store.mongoCollection.Count(ctx, createCollectionContainsNotUploadedFilesQuery(collectionID))

	if err != nil {
		log.Error(ctx, "failed to count unpublishable files", err, log.Data{"collection_id": collectionID})
		return err
	}

	if count > 0 {
		event := fmt.Sprintf("can not publish collection, not all files in %s state", StateUploaded)
		log.Info(ctx, event, log.Data{"collection_id": collectionID, "num_file_not_state_uploaded": count})
		return ErrFileNotInUploadedState
	}

	_, err = store.mongoCollection.UpdateMany(
		ctx,
		bson.M{"collection_id": collectionID},
		bson.D{
			{"$set", bson.D{
				{"state", StatePublished},
				{"last_modified", store.clock.GetCurrentTime()},
				{"published_at", store.clock.GetCurrentTime()}}},
		})

	if err != nil {
		event := fmt.Sprintf("failed to change files to %s state", StatePublished)
		log.Error(ctx, event, err, log.Data{"collection_id": collectionID})
		return err
	}

	col := make([]files.StoredRegisteredMetaData, 0)
	_, err = store.mongoCollection.Find(ctx, bson.M{"collection_id": collectionID}, &col)

	if err != nil {
		return err
	}

	for _, m := range col {
		err = store.kafka.Send(files.AvroSchema, &files.FilePublished{
			Path:        m.Path,
			Etag:        m.Etag,
			Type:        m.Type,
			SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
		})

		if err != nil {
			logdata := log.Data{}
			logdata["metadata"] = m
			log.Error(ctx, "sending published message to kafka", err, logdata)
		}
	}

	return nil
}
