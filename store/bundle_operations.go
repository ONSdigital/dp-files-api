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

func (store *Store) MarkBundlePublished(ctx context.Context, bundleID string) error {
	logdata := log.Data{"bundle_id": bundleID}

	empty, err := store.IsBundleEmpty(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "failed to check if bundle is empty", err, logdata)
		return err
	}
	if empty {
		log.Error(ctx, "bundle empty check fail", ErrNoFilesInBundle, logdata)
		return ErrNoFilesInBundle
	}

	allUploaded, err := store.IsBundleUploaded(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "failed to check if bundle is uploaded", err, logdata)
		return err
	}
	if !allUploaded {
		log.Error(ctx, "bundle uploaded check fail", ErrFileNotInUploadedState, logdata)
		return ErrFileNotInUploadedState
	}

	err = store.updateBundleState(ctx, bundleID, StatePublished)
	if err != nil {
		return err
	}

	requestID := request.GetRequestId(ctx)
	newCtx := request.WithRequestId(context.Background(), requestID)
	go store.NotifyBundlePublished(newCtx, bundleID)

	return nil
}

func (store *Store) IsBundleUploaded(ctx context.Context, bundleID string) (bool, error) {
	published, err := store.IsBundlePublished(ctx, bundleID)
	if err != nil {
		return false, err
	}
	if published {
		return false, nil
	}

	metadata := files.StoredRegisteredMetaData{}
	err = store.metadataCollection.FindOne(ctx, bson.M{"$and": []bson.M{
		{fieldBundleID: bundleID},
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

func (store *Store) IsBundlePublished(ctx context.Context, bundleID string) (bool, error) {
	bundle, err := store.GetBundlePublishedMetadata(ctx, bundleID)
	if err != nil {
		// If there's no record of bundle being published in bundles DB, fall back
		// to the older method that checks the file statuses (if all files in the bundle are marked
		// as published, we consider the bundle published).
		if errors.Is(err, ErrBundleMetadataNotRegistered) {
			return store.AreAllBundleFilesPublished(ctx, bundleID)
		}
		// we've hit an unexpected error
		return false, fmt.Errorf("bundle published check: %w", err)
	}
	if bundle.State == StatePublished {
		return true, nil
	}
	return false, nil
}

func (store *Store) updateBundleState(ctx context.Context, bundleID, state string) error {
	logdata := log.Data{"bundle_id": bundleID, "state": state}

	now := store.clock.GetCurrentTime()

	fields := bson.D{
		{Key: fieldState, Value: state},
		{Key: fieldLastModified, Value: now},
	}

	// TODO: uncomment when PublishedAt is added to StoredBundle struct
	// if state == StatePublished {
	// 	fields = append(fields, bson.E{Key: fieldPublishedAt, Value: now})
	// }

	_, err := store.bundlesCollection.Upsert(
		ctx,
		bson.M{fieldID: bundleID},
		bson.D{
			{Key: "$set", Value: fields},
		})
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to change bundle %v to %s state", bundleID, state), err, logdata)
		return err
	}
	return nil
}

func (store *Store) GetBundlePublishedMetadata(ctx context.Context, id string) (files.StoredBundle, error) {
	bundle := files.StoredBundle{}
	err := store.bundlesCollection.FindOne(ctx, bson.M{fieldID: id}, &bundle)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return files.StoredBundle{}, ErrBundleMetadataNotRegistered
		}
		log.Error(ctx, "bundle metadata fetch error", err, log.Data{"id": id})
		return files.StoredBundle{}, err
	}
	return bundle, err
}

func (store *Store) AreAllBundleFilesPublished(ctx context.Context, bundleID string) (bool, error) {
	empty, err := store.IsBundleEmpty(ctx, bundleID)
	if err != nil {
		return false, fmt.Errorf("AreAllBundleFilesPublished empty bundle check: %w", err)
	}
	if empty {
		return false, nil
	}

	metadata := files.StoredRegisteredMetaData{}
	err = store.metadataCollection.FindOne(ctx, bson.M{"$and": []bson.M{
		{fieldBundleID: bundleID},
		{fieldState: bson.M{"$ne": StatePublished}},
		{fieldState: bson.M{"$ne": StateMoved}},
	}}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return true, nil
		}
		return false, fmt.Errorf("AreAllBundleFilesPublished check: %w", err)
	}
	return false, nil
}

func (store *Store) IsBundleEmpty(ctx context.Context, bundleID string) (bool, error) {
	metadata := files.StoredRegisteredMetaData{}

	err := store.metadataCollection.FindOne(ctx, bson.M{fieldBundleID: bundleID}, &metadata)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return true, nil
		}
		return true, err
	}

	return false, nil
}

func (store *Store) registerBundle(ctx context.Context, bundleID string) error {
	logdata := log.Data{"bundle_id": bundleID}
	now := store.clock.GetCurrentTime()
	bundle := files.StoredBundle{
		ID:           bundleID,
		State:        StateCreated,
		LastModified: now,
	}
	if _, err := store.bundlesCollection.Insert(ctx, bundle); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Info(ctx, "bundle already registered", logdata)
			return nil
		}
		log.Error(ctx, "failed to insert bundle record", err, log.Data{"bundle": config.BundlesCollection, "record": bundle})
		return err
	}
	return nil
}

func (store *Store) UpdateBundleID(ctx context.Context, path, bundleID string) error {
	metadata := files.StoredRegisteredMetaData{}
	logdata := log.Data{"path": path}

	if err := store.metadataCollection.FindOne(ctx, bson.M{"path": path}, &metadata); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "update bundle ID: attempted to operate on unregistered file", err, logdata)
			return ErrFileNotRegistered
		}
		log.Error(ctx, "failed finding metadata to update bundle ID", err, logdata)
		return err
	}

	if bundleID == "" {
		if metadata.BundleID == nil {
			return nil
		}

		_, err := store.metadataCollection.Update(
			ctx,
			bson.M{"path": path},
			bson.D{
				{Key: "$unset", Value: bson.D{
					{Key: "bundle_id", Value: ""},
				}},
			})

		if err != nil {
			log.Error(ctx, "failed to remove bundle ID", err, logdata)
		}
		return err
	}

	if metadata.BundleID != nil {
		logdata["bundle_id"] = *metadata.BundleID
		log.Error(ctx, "update bundle ID: bundle ID already set", ErrBundleIDAlreadySet, logdata)
		return ErrBundleIDAlreadySet
	}

	// check to see if bundleID exists and is not-published
	published, err := store.IsBundlePublished(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "update bundle ID: caught db error", err, logdata)
		return err
	}
	if published {
		log.Error(ctx, fmt.Sprintf("bundle with id [%s] is already published", bundleID), ErrBundleAlreadyPublished, logdata)
		return ErrBundleAlreadyPublished
	}

	_, err = store.metadataCollection.Update(
		ctx,
		bson.M{"path": path},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "bundle_id", Value: bundleID}},
			},
		})

	return err
}

func (store *Store) NotifyBundlePublished(ctx context.Context, bundleID string) {
	// ignoring err as this would have been done previously
	totalCount, _ := store.metadataCollection.Count(ctx, bson.M{fieldBundleID: bundleID})
	log.Info(ctx, "notify bundle published start", log.Data{"bundle_id": bundleID, "total_files": totalCount})
	// balance the number of batches Vs batch size
	batchSize := store.cfg.MinBatchSize
	numBatches := int(math.Ceil(float64(totalCount) / float64(batchSize)))
	if numBatches > store.cfg.MaxNumBatches {
		numBatches = store.cfg.MaxNumBatches
		batchSize = int(math.Ceil(float64(totalCount) / float64(numBatches)))
	}

	var wg sync.WaitGroup
	wg.Add(numBatches)
	for i := 0; i < numBatches; i++ {
		offset := i * batchSize
		cursor, err := store.metadataCollection.FindCursor(ctx, bson.M{fieldBundleID: bundleID}, mongodriver.Offset(offset))
		if err != nil {
			wg.Done()
			log.Error(ctx, "BatchSendKafkaMessages: failed to query collection", err, log.Data{"bundle_id": bundleID})
			continue
		}
		go store.BatchSendBundleKafkaMessages(ctx, cursor, &wg, bundleID, offset, batchSize, i)
	}
	wg.Wait()

	log.Info(ctx, "notify bundle published end", log.Data{"bundle_id": bundleID})
}

func (store *Store) BatchSendBundleKafkaMessages(
	ctx context.Context,
	cursor mongodriver.Cursor,
	wg *sync.WaitGroup,
	bundleID string,
	offset,
	batchSize,
	batchNum int,
) {
	defer wg.Done()
	ld := log.Data{"bundle_id": bundleID, "offset": offset, "batch_size": batchSize, "batch_num": batchNum}
	log.Info(ctx, "BatchSendBundleKafkaMessages", ld)
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error(ctx, "BatchSendBundleKafkaMessages: failed to close cursor", err, ld)
		}
	}()

	for i := 0; i < batchSize; i++ {
		if cursor.Next(ctx) {
			var m files.StoredRegisteredMetaData
			if err := cursor.Decode(&m); err != nil {
				log.Error(ctx, "BatchSendBundleKafkaMessages: failed to decode cursor", err, ld)
				continue
			}
			fp := &files.FilePublished{
				Path:        m.Path,
				Type:        m.Type,
				Etag:        m.Etag,
				SizeInBytes: strconv.FormatUint(m.SizeInBytes, 10),
			}
			if err := store.kafka.Send(files.AvroSchema, fp); err != nil {
				log.Error(ctx, "BatchSendBundleKafkaMessages: can't send message to kafka", err, log.Data{"metadata": m})
			}
		} else {
			break
		}
	}
	if err := cursor.Err(); err != nil {
		log.Error(ctx, "BatchSendBundleKafkaMessages: cursor error", err, ld)
	}

	log.Info(ctx, "BatchSendBundleKafkaMessages end", ld)
}
