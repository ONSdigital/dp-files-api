package store_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestMarkBundlePublishedBundleEmptyCheckReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred during files count")

	bundleCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: BundleFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&bundleCountReturnsError, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to check if bundle is empty", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkBundlePublishedBundleEmpty() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	bundleCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: BundleFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&bundleCountReturnsError, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("bundle empty check fail", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrNoFilesInBundle)
}

func (suite *StoreSuite) TestMarkBundlePublishedWhenFileExistsInStateOtherThanUploaded() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSucceeds(), // there are some files in the bundle
	}
	bundleCollection := mock.MongoCollectionMock{
		FindOneFunc: BundleFindOneSucceeds(), // bundle is not PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, &bundleCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("bundle uploaded check fail", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkBundlePublishedBundleUploadedCheckReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred during uploaded check")

	collectionCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                  // there are some files in the bundle
			{CollectionFindOneReturnsError(expectedError), 1}, // but UPLOADED check fails
		}),
	}
	bundleCollection := mock.MongoCollectionMock{
		FindOneFunc: BundleFindOneSucceeds(), // bundle is not PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsError, nil, &bundleCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to check if bundle is uploaded", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkBundlePublishedBundlePublishedCheckReturnsTrue() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := files.StoredRegisteredMetaData{
		State: store.StateUploaded,
	}
	metadataBytes, _ := bson.Marshal(metadata)
	bundle := files.StoredBundle{
		State: store.StatePublished,
	}
	bundleBytes, _ := bson.Marshal(bundle)

	collectionCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                             // there are some files in the bundle
			{CollectionFindOneSetsResultAndReturnsNil(metadataBytes), 1}, // but the bundle is PUBLISHED
		}),
	}

	bundleCollection := mock.MongoCollectionMock{
		FindOneFunc: BundleFindOneSetsResultAndReturnsNil(bundleBytes), // but the bundle is PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsError, nil, &bundleCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("bundle uploaded check fail", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkBundlePublishedPersistenceFailure() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred")
	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                                   // there are some files in the bundle
			{CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound), 2}, // all of them are UPLOADED
		}),
	}
	bundleColl := mock.MongoCollectionMock{
		UpsertFunc: func(ctx context.Context, selector, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, expectedError
		},
		FindOneFunc: BundleFindOneSucceeds(), // bundle is not PUBLISHED
	}
	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to change bundle 789 to PUBLISHED state", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkBundlePublishedFindCalled() {
	expectedError := errors.New("an error occurred")

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                                   // there are some files in the bundle
			{CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound), 2}, // all of them are UPLOADED
		}),
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(nil, expectedError),
	}
	bundleColl := mock.MongoCollectionMock{
		UpsertFunc: func(ctx context.Context, selector, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
		FindOneFunc: BundleFindOneSucceeds(), // bundle is not PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.Eventually(func() bool {
		return len(metadataColl.FindCursorCalls()) == 1
	}, time.Second, 10*time.Millisecond)
}

func (suite *StoreSuite) TestNotifyBundlePublishedFindErrored() {
	expectedError := errors.New("an error occurred")

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(nil, expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	subject.NotifyBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.Eventually(func() bool {
		return len(collection.FindCursorCalls()) == 1
	}, time.Second, 10*time.Millisecond)
}

func (suite *StoreSuite) TestNotifyBundlePublishedPersistenceSuccess() {
	metadata := suite.generateBundleMetadata(suite.defaultBundleID)
	metadataBytes, _ := bson.Marshal(metadata)

	cursor := mock.MongoCursorMock{
		CloseFunc: func(ctx context.Context) error { return nil },
		NextFunc:  CursorReturnsNumberOfNext(5),
		DecodeFunc: func(val interface{}) error {
			return bson.Unmarshal(metadataBytes, val)
		},
		ErrFunc: func() error { return nil },
	}

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(&cursor, nil),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: func(schema *avro.Schema, event interface{}) error {
			filePublished := event.(*files.FilePublished)

			suite.Equal(metadata.Path, filePublished.Path)
			suite.Equal(metadata.Etag, filePublished.Etag)
			suite.Equal(metadata.Type, filePublished.Type)
			suite.Equal(strconv.FormatUint(metadata.SizeInBytes, 10), filePublished.SizeInBytes)

			return nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &kafkaMock, suite.defaultClock, nil, cfg)

	subject.NotifyBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(5, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestBatchingWithLargeNumberOfFilesBundle() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()
	numFiles := 5000
	cfg, _ := config.Get()
	expectedBatchSize := int(math.Ceil(float64(numFiles) / float64(cfg.MaxNumBatches)))

	metadata := suite.generateBundleMetadata(suite.defaultBundleID)
	metadataBytes, _ := bson.Marshal(metadata)

	cursor := mock.MongoCursorMock{
		CloseFunc: func(ctx context.Context) error { return nil },
		NextFunc:  CursorReturnsNumberOfNext(numFiles),
		DecodeFunc: func(val interface{}) error {
			return bson.Unmarshal(metadataBytes, val)
		},
		ErrFunc: func() error { return nil },
	}

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsValueAndNil(numFiles),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(&cursor, nil),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: func(schema *avro.Schema, event interface{}) error {
			filePublished := event.(*files.FilePublished)

			suite.Equal(metadata.Path, filePublished.Path)
			suite.Equal(metadata.Etag, filePublished.Etag)
			suite.Equal(metadata.Type, filePublished.Type)
			suite.Equal(strconv.FormatUint(metadata.SizeInBytes, 10), filePublished.SizeInBytes)

			return nil
		},
	}

	subject := store.NewStore(&collection, nil, nil, &kafkaMock, suite.defaultClock, nil, cfg)

	subject.NotifyBundlePublished(suite.defaultContext, suite.defaultBundleID)

	evts := suite.logInterceptor.GetLogEvents("BatchSendBundleKafkaMessages")

	for _, evt := range evts {
		suite.EqualValues(evt["batch_size"].(float64), expectedBatchSize)
	}
	// make sure correct number of messages are sent
	suite.Equal(len(kafkaMock.SendCalls()), numFiles)
	suite.Equal(cfg.MaxNumBatches, len(evts))
	suite.Equal(cfg.MaxNumBatches, len(cursor.ErrCalls()))
	suite.Equal(cfg.MaxNumBatches, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestNotifyBundlePublishedKafkaErrorDoesNotFailOperation() {
	metadata := suite.generateBundleMetadata(suite.defaultBundleID)
	metadataBytes, _ := bson.Marshal(metadata)

	kafkaError := errors.New("an error occurred with Kafka")

	cursor := mock.MongoCursorMock{
		CloseFunc: func(ctx context.Context) error { return nil },
		NextFunc:  CursorReturnsNumberOfNext(5),
		DecodeFunc: func(val interface{}) error {
			return bson.Unmarshal(metadataBytes, val)
		},
		ErrFunc: func() error { return nil },
	}
	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(&cursor, nil),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsError(kafkaError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &kafkaMock, suite.defaultClock, nil, cfg)

	subject.NotifyBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(5, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestNotifyBundlePublishedDecodeErrorDoesNotFailOperation() {
	cursor := mock.MongoCursorMock{
		CloseFunc: func(ctx context.Context) error { return nil },
		NextFunc:  CursorReturnsNumberOfNext(5),
		DecodeFunc: func(val interface{}) error {
			return errors.New("decode error")
		},
		ErrFunc: func() error { return nil },
	}
	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(&cursor, nil),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsError(errors.New("an error occurred with Kafka")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &kafkaMock, suite.defaultClock, nil, cfg)

	subject.NotifyBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(0, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

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

func (suite *StoreSuite) TestUpdateBundleIDFileIsInMovedState() {
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

	suite.Equal("update bundle ID: attempted to operate on a moved file", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileMoved)
}

func (suite *StoreSuite) TestUpdateBundleIDtoEmptyStringFileIsInMovedState() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateBundleMetadata(suite.defaultBundleID)
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, "")
	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("update bundle ID: attempted to operate on a moved file", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileMoved)
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

func (suite *StoreSuite) TestUpdateBundleIDRemoveBundleID() {
	metadata := suite.generateBundleMetadata("existing-bundle-id")
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithFileHavingBundleID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithFileHavingBundleID, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, "")

	suite.NoError(err)
	suite.Equal(1, len(collectionWithFileHavingBundleID.UpdateCalls()))
}

func (suite *StoreSuite) TestUpdateBundleIDRemoveBundleIDNoExistingBundle() {
	metadata := suite.generateBundleMetadata("")
	metadata.State = store.StateUploaded
	metadata.BundleID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithFileHavingNoBundleID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithFileHavingNoBundleID, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, "")

	suite.NoError(err)
	suite.Equal(0, len(collectionWithFileHavingNoBundleID.UpdateCalls()))
}

func (suite *StoreSuite) TestUpdateBundleIDRemoveBundleIDUpdateError() {
	metadata := suite.generateBundleMetadata("existing-bundle-id")
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred during update")

	collectionWithFileHavingBundleID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithFileHavingBundleID, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateBundleID(suite.defaultContext, suite.path, "")

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}
