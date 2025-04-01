package store_test

import (
	"context"
	"errors"
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

func (suite *StoreSuite) TestUpdateCollectionIDFindReturnsErrNoDocumentFound() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("update collection ID: attempted to operate on unregistered file", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestUpdateCollectionIDFindReturnsUnspecifiedError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, "", suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed finding metadata to update collection ID", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDCollectionIDAlreadySet() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)
	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("update collection ID: collection ID already set", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrCollectionIDAlreadySet)
}

func (suite *StoreSuite) TestUpdateCollectionIDCollectionCheckFail() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateCollectionMetadata("")
	metadata.State = store.StateUploaded
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSetsResultAndReturnsNil(metadataBytes), 1},
		}),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)

	suite.Equal(true, suite.logInterceptor.IsEventPresent("update collection ID: caught db error"))
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDCollectionAlreadyPublished() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateCollectionMetadata("")
	metadata.State = store.StatePublished
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	coll, _ := bson.Marshal(files.StoredCollection{
		State: store.StatePublished,
	})
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(coll), // collection is PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)
	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection with id [123456] is already published", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrCollectionAlreadyPublished)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateReturnsError() {
	metadata := suite.generateCollectionMetadata("")
	metadata.State = store.StateUploaded
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &emptyCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateSuccess() {
	metadata := suite.generateCollectionMetadata("")
	metadata.State = store.StateUploaded
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionContainsOneUploadedFileWithNoCollectionID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionContainsOneUploadedFileWithNoCollectionID, &emptyCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCollectionEmptyCheckReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred during files count")

	collectionCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsError, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to check if collection is empty", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCollectionEmpty() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsError, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection empty check fail", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrNoFilesInCollection)
}

func (suite *StoreSuite) TestMarkCollectionPublishedWhenFileExistsInStateOtherThanUploaded() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSucceeds(), // there are some files in the collection
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSucceeds(), // collection is not PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection uploaded check fail", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCollectionUploadedCheckReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred during uploaded check")

	collectionCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                  // there are some files in the collection
			{CollectionFindOneReturnsError(expectedError), 1}, // but UPLOADED check fails
		}),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSucceeds(), // collection is not PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsError, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to check if collection is uploaded", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}
func (suite *StoreSuite) TestMarkCollectionPublishedCollectionPublishedCheckReturnsTrue() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := files.StoredRegisteredMetaData{
		State: store.StateUploaded,
	}
	metadataBytes, _ := bson.Marshal(metadata)
	collection := files.StoredCollection{
		State: store.StatePublished,
	}
	collectionBytes, _ := bson.Marshal(collection)

	collectionCountReturnsError := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                             // there are some files in the collection
			{CollectionFindOneSetsResultAndReturnsNil(metadataBytes), 1}, // but the collection is PUBLISHED
		}),
	}

	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(collectionBytes), // but the collection is PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsError, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection uploaded check fail", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkCollectionPublishedPersistenceFailure() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred")
	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                                   // there are some files in the collection
			{CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound), 2}, // all of them are UPLOADED
		}),
	}
	collectionColl := mock.MongoCollectionMock{
		UpsertFunc: func(ctx context.Context, selector, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, expectedError
		},
		FindOneFunc: CollectionFindOneSucceeds(), // collection is not PUBLISHED
	}
	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to change collection 123456 to PUBLISHED state", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedFindCalled() {
	expectedError := errors.New("an error occurred")

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSucceeds(), 1},                                   // there are some files in the collection
			{CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound), 2}, // all of them are UPLOADED
		}),
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(nil, expectedError),
	}
	collectionColl := mock.MongoCollectionMock{
		UpsertFunc: func(ctx context.Context, selector, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
		FindOneFunc: CollectionFindOneSucceeds(), // collection is not PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.NoError(err)
	suite.Eventually(func() bool {
		return len(metadataColl.FindCursorCalls()) == 1
	}, time.Second, 10*time.Millisecond)
}

func (suite *StoreSuite) TestNotifyCollectionPublishedFindErrored() {
	expectedError := errors.New("an error occurred")

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(nil, expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.Eventually(func() bool {
		return len(collection.FindCursorCalls()) == 1
	}, time.Second, 10*time.Millisecond)
}

func (suite *StoreSuite) TestNotifyCollectionPublishedPersistenceSuccess() {
	metadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
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

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(5, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestBatchingWithLargeNumberOfFiles() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()
	numFiles := 5000
	cfg, _ := config.Get()
	expectedBatchSize := int(math.Ceil(float64(numFiles) / float64(cfg.MaxNumBatches)))

	metadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
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

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	evts := suite.logInterceptor.GetLogEvents("BatchSendKafkaMessages")

	for _, evt := range evts {
		suite.EqualValues(evt["batch_size"].(float64), expectedBatchSize)
	}
	// make sure correct number of messages are sent
	suite.Equal(len(kafkaMock.SendCalls()), numFiles)
	suite.Equal(cfg.MaxNumBatches, len(evts))
	suite.Equal(cfg.MaxNumBatches, len(cursor.ErrCalls()))
	suite.Equal(cfg.MaxNumBatches, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestNotifyCollectionPublishedKafkaErrorDoesNotFailOperation() {
	metadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
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

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(5, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestNotifyCollectionPublishedDecodeErrorDoesNotFailOperation() {
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

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(0, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestIsCollectionPublishedNoMetadata() {
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&emptyCollection, &emptyCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.False(published)
	suite.NoError(err)
}

func (suite *StoreSuite) TestIsCollectionPublishedMetadataError() {
	expectedError := errors.New("an error occurred")
	erroringCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, &erroringCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.False(published)
	suite.Error(err)
	suite.EqualError(err, "collection published check: an error occurred")
}

func (suite *StoreSuite) TestIsCollectionPublishedMetadataPublished() {
	metadata := files.StoredRegisteredMetaData{
		State: store.StatePublished,
	}
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneChain([]CollectionFindOneFuncChainEntry{
			{CollectionFindOneSetsResultAndReturnsNil(metadataBytes), 1},       // collection not empty
			{CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound), 1}, // no files found that weren't published
		}),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(store.ErrCollectionMetadataNotRegistered),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.True(published)
	suite.NoError(err)
}

func (suite *StoreSuite) TestIsCollectionPublishedNoCollectionAllFilesNotPublished() {
	metadata := files.StoredRegisteredMetaData{
		State: store.StateUploaded,
	}
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(store.ErrCollectionMetadataNotRegistered),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.False(published)
	suite.NoError(err)
}

func (suite *StoreSuite) TestIsCollectionPublishedCollectionError() {
	metadata := files.StoredRegisteredMetaData{
		State: store.StateUploaded,
	}
	metadataBytes, _ := bson.Marshal(metadata)
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.False(published)
	suite.Error(err)
	suite.EqualError(err, "collection published check: an error occurred")
}

func (suite *StoreSuite) TestIsCollectionPublishedCollectionPublished() {
	metadata := files.StoredRegisteredMetaData{
		State: store.StateUploaded,
	}
	metadataBytes, _ := bson.Marshal(metadata)
	collection := files.StoredCollection{
		State: store.StatePublished,
	}
	collectionBytes, _ := bson.Marshal(collection)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(collectionBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.True(published)
	suite.NoError(err)
}

func (suite *StoreSuite) TestIsCollectionPublishedCollectionNotPublished() {
	metadata := files.StoredRegisteredMetaData{
		State: store.StateUploaded,
	}
	metadataBytes, _ := bson.Marshal(metadata)
	collection := files.StoredCollection{
		State: "",
	}
	collectionBytes, _ := bson.Marshal(collection)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(collectionBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &collCollection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	published, err := subject.IsCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.False(published)
	suite.NoError(err)
}
