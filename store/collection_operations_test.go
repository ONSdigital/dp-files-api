package store_test

import (
	"context"
	"errors"
	"strconv"
	"time"

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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)

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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.UpdateCollectionID(suite.defaultContext, "", suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed finding metadata to update collection ID", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDCollectionIDAlreadySet() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)
	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("update collection ID: collection ID already set", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrCollectionIDAlreadySet)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateReturnsError() {
	metadata := suite.generateMetadata("")
	metadata.State = store.StateUploaded
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateSuccess() {
	metadata := suite.generateMetadata("")
	metadata.State = store.StateUploaded
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionContainsOneUploadedFileWithNoCollectionID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionContainsOneUploadedFileWithNoCollectionID, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.UpdateCollectionID(suite.defaultContext, suite.path, suite.defaultCollectionID)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCountReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred during files count")

	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndError(0, expectedError),
	}

	subject := store.NewStore(&collectionCountReturnsError, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to count files collection", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCountReturnsZero() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(0),
	}

	subject := store.NewStore(&collectionCountReturnsError, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("no files found in collection", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrNoFilesInCollection)
}

func (suite *StoreSuite) TestMarkCollectionPublishedWhenFileExistsInStateOtherThanUploaded() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(1),
	}

	subject := store.NewStore(&collectionCountReturnsError, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("can not publish collection, not all files in UPLOADED state", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkCollectionPublishedPersistenceFailure() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred")
	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collection, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to change files to PUBLISHED state", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedFindCalled() {

	expectedError := errors.New("an error occurred")

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(nil, expectedError),
	}

	subject := store.NewStore(&collection, &suite.defaultKafkaProducer, suite.defaultClock)

	err := subject.MarkCollectionPublished(suite.defaultContext, suite.defaultCollectionID)

	suite.NoError(err)
	suite.Eventually(func() bool {
		return len(collection.FindCursorCalls()) == 1
	}, time.Second, 10*time.Millisecond)
}

func (suite *StoreSuite) TestNotifyCollectionPublishedFindErrored() {

	expectedError := errors.New("an error occurred")

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindCursorFunc: CollectionFindCursorReturnsCursorAndError(nil, expectedError),
	}

	subject := store.NewStore(&collection, &suite.defaultKafkaProducer, suite.defaultClock)

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID, 1)

	suite.Eventually(func() bool {
		return len(collection.FindCursorCalls()) == 1
	}, time.Second, 10*time.Millisecond)
}

func (suite *StoreSuite) TestNotifyCollectionPublishedPersistenceSuccess() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
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

	subject := store.NewStore(&collection, &kafkaMock, suite.defaultClock)

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID, 5)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(5, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}

func (suite *StoreSuite) TestNotifyCollectionPublishedKafkaErrorDoesNotFailOperation() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
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

	subject := store.NewStore(&collection, &kafkaMock, suite.defaultClock)

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID, 5)

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

	subject := store.NewStore(&collection, &kafkaMock, suite.defaultClock)

	subject.NotifyCollectionPublished(suite.defaultContext, suite.defaultCollectionID, 5)

	suite.Equal(6, len(cursor.NextCalls()))
	suite.Equal(5, len(cursor.DecodeCalls()))
	suite.Equal(0, len(kafkaMock.SendCalls()))
	suite.Equal(1, len(cursor.ErrCalls()))
	suite.Equal(1, len(cursor.CloseCalls()))
}
