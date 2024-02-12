package store_test

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (suite *StoreSuite) TestRegisterFileUploadCollectionPublishedCheckError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StatePublished

	expectedError := errors.New("collection fetch error")
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, &collCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.Equal(true, suite.logInterceptor.IsEventPresent("collection published check error"))
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenCollectionAlreadyPublished() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StatePublished

	coll, _ := bson.Marshal(files.StoredCollection{
		State: store.StatePublished,
	})
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(coll), // collection is PUBLISHED
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, &collCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection is already published", logEvent)
	suite.ErrorIs(err, store.ErrCollectionAlreadyPublished)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathAlreadyExists() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	expectedError := mongo.WriteError{Code: 11000}
	metadataBytes, _ := bson.Marshal(metadata)

	alwaysFindsExistingCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc:  CollectionInsertReturnsNilAndError(expectedError),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&alwaysFindsExistingCollection, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("file upload already registered", logEvent)
	suite.ErrorIs(err, store.ErrDuplicateFile)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFileDoesNotAlreadyExist() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = ""
	metadataBytes, _ := bson.Marshal(metadata)

	collectionCountReturnsZero := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			actualMetadata := document.(files.StoredRegisteredMetaData)

			testCurrentTime := suite.defaultClock.GetCurrentTime()

			suite.Equal(store.StateCreated, actualMetadata.State)
			suite.Equal(testCurrentTime, actualMetadata.CreatedAt)
			suite.Equal(testCurrentTime, actualMetadata.LastModified)

			suite.assertImmutableFieldsUnchanged(metadata, actualMetadata)

			return nil, nil
		},
	}
	coll, _ := bson.Marshal(files.StoredCollection{
		State: store.StateUploaded,
	})
	collCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(coll), // collection is not PUBLISHED
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			actualCollection := document.(files.StoredCollection)
			testCurrentTime := suite.defaultClock.GetCurrentTime()

			suite.Equal(store.StateCreated, actualCollection.State)
			suite.Equal(testCurrentTime, actualCollection.LastModified)
			suite.Equal(suite.defaultCollectionID, actualCollection.ID)

			return nil, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsZero, &collCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.NoError(err)
}

func (suite *StoreSuite) TestRegisterFileUploadInsertReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	expectedError := errors.New("error occurred")
	metadataBytes, _ := bson.Marshal(metadata)

	collectionCountReturnsZero := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc:  CollectionInsertReturnsNilAndError(expectedError),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsZero, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to insert metadata", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestRegisterFileUploadInsertSucceeds() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	collectionCountReturnsZero := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc:  CollectionInsertReturnsNilAndNil(),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
		InsertFunc: CollectionInsertReturnsNilAndNil(),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsZero, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("registering new file upload", logEvent)
	suite.NoError(err)
}

func (suite *StoreSuite) TestRegisterFileUploadRegisterCollectionFails() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	collectionCountReturnsZero := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc:  CollectionInsertReturnsNilAndNil(),
	}
	expectedError := errors.New("collection insert error")
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
		InsertFunc: CollectionInsertReturnsNilAndError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionCountReturnsZero, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.Equal(true, suite.logInterceptor.IsEventPresent("failed to register collection"))
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenNotInCreatedState() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)

	tests := []struct {
		currentState string
		expectedErr  error
	}{
		{store.StateUploaded, store.ErrFileStateMismatch},
		{store.StatePublished, store.ErrFileStateMismatch},
	}

	for _, test := range tests {
		metadata.State = test.currentState

		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		}
		emptyCollection := mock.MongoCollectionMock{
			FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
				return mongodriver.ErrNoDocumentFound
			},
		}

		cfg, _ := config.Get()
		subject := store.NewStore(&collectionWithUploadedFile, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
		err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

		logEvents := suite.logInterceptor.GetLogEvents("update file state: state mismatch")

		suite.Equal(1, len(logEvents))
		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenFileNotExists() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StateCreated

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

	logEvents := suite.logInterceptor.GetLogEvents("update file state: attempted to operate on unregistered file")

	suite.Equal(1, len(logEvents))
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkUploadCompleteSucceeds() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateManyReturnsNilAndNil(),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsErrNoDocumentFound() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	logEvents := suite.logInterceptor.GetLogEvents("mark file as published: attempted to operate on unregistered file")

	suite.Equal(1, len(logEvents))
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsUnspecifiedError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("mark file as published: failed finding file metadata", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedCollectionIDNil() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("file had no collection id", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrCollectionIDNotSet)
}

func (suite *StoreSuite) TestMarkFilePublishedStateUploaded() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	notUploadedStates := []string{
		store.StateCreated,
		store.StatePublished,
	}

	for _, state := range notUploadedStates {
		metadata := suite.generateMetadata(suite.defaultCollectionID)
		metadata.State = state
		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		}

		cfg, _ := config.Get()
		subject := store.NewStore(&collectionWithUploadedFile, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
		err := subject.MarkFilePublished(suite.defaultContext, suite.path)

		logEvent := suite.logInterceptor.GetLogEvent()

		suite.Equal("mark file published: file was not in state UPLOADED", logEvent)
		suite.Error(err)
		suite.ErrorIs(err, store.ErrFileNotInUploadedState, "%s is not %s", state, store.StateUploaded)
	}
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
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
	subject := store.NewStore(&collectionWithUploadedFile, &emptyCollection, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateKafkaReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &emptyCollection, &kafkaMock, suite.defaultClock, nil, cfg)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateKafkaDoesNotReturnError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}
	emptyCollection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsNil(),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collectionWithUploadedFile, &emptyCollection, &kafkaMock, suite.defaultClock, nil, cfg)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.NoError(err)
}
