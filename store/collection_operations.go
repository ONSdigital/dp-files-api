package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (store *Store) IsCollectionPublished(ctx context.Context, collectionID string) (bool, error) {
	coll, err := store.GetCollectionPublishedMetadata(ctx, collectionID)
	if err != nil {
		// If there's no record of collection being published in collections DB, fall back
		// to the older method that checks the file statuses (if all files in the collection are marked
		// as published, we consider the collection published).
		if errors.Is(err, ErrCollectionMetadataNotRegistered) {
			return store.AreAllFilesPublished(ctx, collectionID)
		}
		// we've hit an unexpected error
		return false, fmt.Errorf("collection published check: %w", err)
	}
	if coll.State == StatePublished {
		return true, nil
	}
	return false, nil
}

func (store *Store) AreAllFilesPublished(ctx context.Context, collectionID string) (bool, error) {
	empty, err := store.IsCollectionEmpty(ctx, collectionID)
	if err != nil {
		return false, fmt.Errorf("AreAllFilesPublished empty collection check: %w", err)
	}
	if empty {
		return false, nil
	}

	metadata := files.StoredRegisteredMetaData{}
	err = store.metadataCollection.FindOne(ctx, bson.M{"$and": []bson.M{
		{fieldCollectionID: collectionID},
		{fieldState: bson.M{"$ne": StatePublished}},
	}}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return true, nil
		}
		return false, fmt.Errorf("AreAllFilesPublished check: %w", err)
	}
	return false, nil
}

func (store *Store) UpdateCollectionID(ctx context.Context, path, collectionID string) error {
	metadata := files.StoredRegisteredMetaData{}
	logdata := log.Data{"path": path}

	if err := store.metadataCollection.FindOne(ctx, bson.M{"path": path}, &metadata); err != nil {
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

	//check to see if collectionID exists and is not-published
	published, err := store.IsCollectionPublished(ctx, collectionID)
	if err != nil {
		log.Error(ctx, "update collection ID: caught db error", err, logdata)
		return err
	}
	if published {
		log.Error(ctx, fmt.Sprintf("collection with id [%s] is already published", collectionID), ErrCollectionAlreadyPublished, logdata)
		return ErrCollectionAlreadyPublished
	}

	_, err = store.metadataCollection.Update(
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
	logdata := log.Data{"collection_id": collectionID}

	empty, err := store.IsCollectionEmpty(ctx, collectionID)
	if err != nil {
		log.Error(ctx, "failed to check if collection is empty", err, logdata)
		return err
	}
	if empty {
		log.Error(ctx, "collection empty check fail", ErrNoFilesInCollection, logdata)
		return ErrNoFilesInCollection
	}

	allUploaded, err := store.IsCollectionUploaded(ctx, collectionID)
	if err != nil {
		log.Error(ctx, "failed to check if collection is uploaded", err, logdata)
		return err
	}
	if !allUploaded {
		log.Error(ctx, "collection uploaded check fail", ErrFileNotInUploadedState, logdata)
		return ErrFileNotInUploadedState
	}

	err = store.updateCollectionState(ctx, collectionID, StatePublished)
	if err != nil {
		return err
	}

	requestID := request.GetRequestId(ctx)
	newCtx := request.WithRequestId(context.Background(), requestID)
	go store.NotifyCollectionPublished(newCtx, collectionID)

	return nil
}

func (store *Store) updateCollectionState(ctx context.Context, collectionID string, state string) error {
	logdata := log.Data{"collection_id": collectionID, "state": state}

	now := store.clock.GetCurrentTime()

	fields := bson.D{
		{fieldState, state},
		{fieldLastModified, now},
	}

	if state == StatePublished {
		fields = append(fields, bson.E{fieldPublishedAt, now})
	}

	_, err := store.collectionsCollection.Upsert(
		ctx,
		bson.M{fieldID: collectionID},
		bson.D{
			{"$set", fields},
		})
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to change collection %v to %s state", collectionID, state), err, logdata)
		return err
	}
	return nil
}
func (store *Store) registerCollection(ctx context.Context, collectionID string) error {
	logdata := log.Data{"collection_id": collectionID}
	now := store.clock.GetCurrentTime()
	collection := files.StoredCollection{
		ID:           collectionID,
		State:        StateCreated,
		LastModified: now,
	}
	if _, err := store.collectionsCollection.Insert(ctx, collection); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Info(ctx, "collection already registered", logdata)
			return nil
		}
		log.Error(ctx, "failed to insert collection record", err, log.Data{"collection": config.CollectionsCollection, "record": collection})
		return err
	}
	return nil
}

func (store *Store) IsCollectionEmpty(ctx context.Context, collectionID string) (bool, error) {
	metadata := files.StoredRegisteredMetaData{}

	err := store.metadataCollection.FindOne(ctx, bson.M{fieldCollectionID: collectionID}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return true, nil
		}
		return true, err
	}

	return false, nil
}

func (store *Store) IsCollectionUploaded(ctx context.Context, collectionID string) (bool, error) {
	published, err := store.IsCollectionPublished(ctx, collectionID)
	if err != nil {
		return false, err
	}
	if published {
		return false, nil
	}

	metadata := files.StoredRegisteredMetaData{}
	err = store.metadataCollection.FindOne(ctx, bson.M{"$and": []bson.M{
		{fieldCollectionID: collectionID},
		{fieldState: bson.M{"$ne": StateUploaded}},
	}}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (store *Store) NotifyCollectionPublished(ctx context.Context, collectionID string) {
	// ignoring err as this would have been done previously
	totalCount, _ := store.metadataCollection.Count(ctx, bson.M{fieldCollectionID: collectionID})
	log.Info(ctx, "notify collection published start", log.Data{"collection_id": collectionID, "total_files": totalCount})
	// balance the number of batches Vs batch size
	batch_size := store.cfg.MinBatchSize
	num_batches := int(math.Ceil(float64(totalCount) / float64(batch_size)))
	if num_batches > store.cfg.MaxNumBatches {
		num_batches = store.cfg.MaxNumBatches
		batch_size = int(math.Ceil(float64(totalCount) / float64(num_batches)))
	}

	var wg sync.WaitGroup
	wg.Add(num_batches)
	for i := 0; i < num_batches; i++ {
		offset := i * batch_size
		cursor, err := store.metadataCollection.FindCursor(ctx, bson.M{fieldCollectionID: collectionID}, mongodriver.Offset(offset))
		if err != nil {
			wg.Done()
			log.Error(ctx, "BatchSendKafkaMessages: failed to query collection", err, log.Data{"collection_id": collectionID})
			continue
		}
		go store.BatchSendKafkaMessages(ctx, cursor, &wg, collectionID, offset, batch_size, i)
	}
	wg.Wait()

	log.Info(ctx, "notify collection published end", log.Data{"collection_id": collectionID})
}

func (store *Store) BatchSendKafkaMessages(
	ctx context.Context,
	cursor mongodriver.Cursor,
	wg *sync.WaitGroup,
	collectionID string,
	offset,
	batch_size,
	batch_num int,
) {
	defer wg.Done()
	ld := log.Data{"collection_id": collectionID, "offset": offset, "batch_size": batch_size, "batch_num": batch_num}
	log.Info(ctx, "BatchSendKafkaMessages", ld)
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error(ctx, "BatchSendKafkaMessages: failed to close cursor", err, ld)
		}
	}()

	for i := 0; i < batch_size; i++ {
		if cursor.Next(ctx) {
			var m files.StoredRegisteredMetaData
			if err := cursor.Decode(&m); err != nil {
				log.Error(ctx, "BatchSendKafkaMessages: failed to decode cursor", err, ld)
				continue
			}
			fp := &files.FilePublished{
				Path:        m.Path,
				Type:        m.Type,
				Etag:        m.Etag,
				SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
			}
			if err := store.kafka.Send(files.AvroSchema, fp); err != nil {
				log.Error(ctx, "BatchSendKafkaMessages: can't send message to kafka", err, log.Data{"metadata": m})
			}
		} else {
			break
		}
	}
	if err := cursor.Err(); err != nil {
		log.Error(ctx, "BatchSendKafkaMessages: cursor error", err, ld)
	}

	log.Info(ctx, "BatchSendKafkaMessages end", ld)
}
