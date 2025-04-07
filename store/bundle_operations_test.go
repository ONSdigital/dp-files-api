package store_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestGetBundlePublishedMetadataSuccess() {
	expectedBundle := files.StoredBundle{
		ID:    suite.defaultBundleID,
		State: store.StatePublished,
	}
	bundleBytes, _ := bson.Marshal(expectedBundle)

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	actualBundle, err := subject.GetBundlePublishedMetadata(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.Exactly(expectedBundle, actualBundle)
}

func (suite *StoreSuite) TestGetBundlePublishedMetadataNotFound() {
	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	actualBundle, err := subject.GetBundlePublishedMetadata(suite.defaultContext, "non-existent-bundle-id")

	suite.Error(err)
	suite.ErrorIs(err, store.ErrBundleMetadataNotRegistered)
	suite.Equal(files.StoredBundle{}, actualBundle)
}

func (suite *StoreSuite) TestGetBundlePublishedMetadataUnexpectedError() {
	expectedError := errors.New("unexpected error")

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	actualBundle, err := subject.GetBundlePublishedMetadata(suite.defaultContext, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.Equal(files.StoredBundle{}, actualBundle)
}
func (suite *StoreSuite) TestIsBundlePublishedSuccess() {
	expectedBundle := files.StoredBundle{
		ID:    suite.defaultBundleID,
		State: store.StatePublished,
	}
	bundleBytes, _ := bson.Marshal(expectedBundle)

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.True(isPublished)
}

func (suite *StoreSuite) TestIsBundlePublishedNotPublishedState() {
	expectedBundle := files.StoredBundle{
		ID:    suite.defaultBundleID,
		State: store.StateCreated,
	}
	bundleBytes, _ := bson.Marshal(expectedBundle)

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestIsBundlePublishedFallbackToFileCheckNotPublished() {
	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(store.ErrBundleMetadataNotRegistered),
	}

	metadata := bson.M{
		"state": store.StateCreated,
	}
	metadataBytes, _ := bson.Marshal(metadata)

	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestIsBundlePublishedUnexpectedError() {
	expectedError := errors.New("unexpected error")

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestAreAllBundleFilesPublishedNotAllFilesPublished() {
	metadata := files.StoredRegisteredMetaData{
		BundleID: &suite.defaultBundleID,
		State:    store.StateCreated,
	}
	metadataBytes, _ := bson.Marshal(metadata)

	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	bundlesCollection := mock.MongoCollectionMock{}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.AreAllBundleFilesPublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestAreAllBundleFilesPublishedEmptyBundle() {
	bundlesCollection := mock.MongoCollectionMock{}
	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.AreAllBundleFilesPublished(suite.defaultContext, "empty-bundle-id")

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestAreAllBundleFilesPublishedUnexpectedError() {
	expectedError := errors.New("unexpected error")

	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	bundlesCollection := mock.MongoCollectionMock{}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.AreAllBundleFilesPublished(suite.defaultContext, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestUpdateBundleIDFindReturnsErrNoDocumentFound() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("update bundle ID: attempted to operate on unregistered file", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestUpdateBundleIDFindReturnsUnspecifiedError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, "", suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed finding metadata to update bundle ID", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateBundleIDBundleIDAlreadySet() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateBundleMetadata(suite.defaultBundleID)
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, suite.defaultBundleID)
	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("update bundle ID: bundle ID already set", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrBundleIDAlreadySet)
}

func (suite *StoreSuite) TestUpdateBundleIDBundleCheckFail() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateBundleMetadata("")
	metadata.State = store.StateUploaded
	metadata.BundleID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSetsResultAndReturnsNil(metadataBytes), 1},
		}),
	}
	bundleCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &bundleCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, suite.defaultBundleID)

	suite.Equal(true, suite.logInterceptor.IsEventPresent("update bundle ID: caught db error"))
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateBundleIDBundleAlreadyPublished() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateBundleMetadata("")
	metadata.State = store.StatePublished
	metadata.BundleID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	bundle, _ := bson.Marshal(files.StoredBundle{
		State: store.StatePublished,
	})
	bundleCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundle),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &bundleCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, suite.defaultBundleID)
	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal(fmt.Sprintf("bundle with id [%s] is already published", suite.defaultBundleID), logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrBundleAlreadyPublished)
}

func (suite *StoreSuite) TestUpdateBundleIDUpdateReturnsError() {
	metadata := suite.generateBundleMetadata("")
	metadata.State = store.StateUploaded
	metadata.BundleID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}
	emptyBundle := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &emptyBundle, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateBundleIDUpdateSuccess() {
	metadata := suite.generateBundleMetadata("")
	metadata.State = store.StateUploaded
	metadata.BundleID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionContainsOneUploadedFileWithNoBundleID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}
	emptyBundle := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionContainsOneUploadedFileWithNoBundleID, nil, &emptyBundle, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, suite.defaultBundleID)

	suite.NoError(err)
}
