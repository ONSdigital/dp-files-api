package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
