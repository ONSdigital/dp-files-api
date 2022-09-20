package store_test

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (suite *StoreSuite) TestRegisterFileUploadWhenCollectionAlreadyPublished() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)
	
	alwaysFindsExistingCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	subject := store.NewStore(&alwaysFindsExistingCollection, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection with id [123456] is already published", logEvent)
	suite.ErrorIs(err, store.ErrCollectionAlreadyPublished)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathAlreadyExists() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	expectedError := mongo.WriteError{Code: 11000}
	metadataBytes, _ := bson.Marshal(metadata)

	alwaysFindsExistingCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc: CollectionInsertReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&alwaysFindsExistingCollection, &suite.defaultKafkaProducer, suite.defaultClock)
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

	subject := store.NewStore(&collectionCountReturnsZero, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.NoError(err)
}

func (suite *StoreSuite) TestRegisterFileUploadInsertReturnsError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	expectedError := errors.New("error occurred")
	metadataBytes, _ := bson.Marshal(metadata)

	collectionCountReturnsZero := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc:  CollectionInsertReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionCountReturnsZero, &suite.defaultKafkaProducer, suite.defaultClock)
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
	metadataBytes, _ := bson.Marshal(metadata)

	collectionCountReturnsZero := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		InsertFunc:  CollectionInsertReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionCountReturnsZero, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("registering new file upload", logEvent)
	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenNotInCreatedState() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)

	tests := []struct {
		currentState string
		expectedErr  error
	}{
		{store.StateUploaded, store.ErrFileNotInPublishedState},
		{store.StatePublished, store.ErrFileNotInPublishedState},
		{store.StateDecrypted, store.ErrFileNotInPublishedState},
	}

	for _, test := range tests {
		metadata.State = test.currentState

		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		}

		subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
		err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

		logEvent := suite.logInterceptor.GetLogEvent()

		suite.Equal("mark file decrypted: file was not in state CREATED", logEvent)
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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("mark file as decrypted: attempted to operate on unregistered file", logEvent)
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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkUploadComplete(suite.defaultContext, suite.etagReference(metadata))

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenNotInCreatedState() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)

	tests := []struct {
		currentState string
		expectedErr  error
	}{
		{store.StateCreated, store.ErrFileNotInPublishedState},
		{store.StateUploaded, store.ErrFileNotInPublishedState},
		{store.StateDecrypted, store.ErrFileNotInPublishedState},
	}

	for _, test := range tests {
		metadata.State = test.currentState

		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		}

		subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
		err := subject.MarkFileDecrypted(suite.defaultContext, suite.etagReference(metadata))

		logEvent := suite.logInterceptor.GetLogEvent()

		suite.Equal("mark file decrypted: file was not in state CREATED", logEvent)
		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenFileNotExists() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StatePublished

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkFileDecrypted(suite.defaultContext, suite.etagReference(metadata))

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("mark file as decrypted: attempted to operate on unregistered file", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenUpdateReturnsError() {

	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkFileDecrypted(suite.defaultContext, suite.etagReference(metadata))

	suite.Error(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedSucceeds() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkFileDecrypted(suite.defaultContext, suite.etagReference(metadata))

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsErrNoDocumentFound() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("mark file as published: attempted to operate on unregistered file", logEvent)
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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed finding metadata to mark file as published", logEvent)
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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
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
		store.StateDecrypted,
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

		subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)
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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultKafkaProducer, suite.defaultClock)

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

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &kafkaMock, suite.defaultClock)

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

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &kafkaMock, suite.defaultClock)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.NoError(err)
}
