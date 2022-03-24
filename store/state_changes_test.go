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
)

func (suite *StoreSuite) TestRegisterFileUploadCountReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	expectedError := errors.New("error occurred")

	AlwaysFindsExistingCollection := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndError(0, expectedError),
	}

	store := store.NewStore(&AlwaysFindsExistingCollection, &suite.defaultkafkaProducer, suite.defaultClock)

	err := store.RegisterFileUpload(suite.defaultContext, metadata)

	suite.Error(err)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathAlreadyExists() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	AlwaysFindsExistingCollection := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(1),
	}

	subject := store.NewStore(&AlwaysFindsExistingCollection, &suite.defaultkafkaProducer, suite.defaultClock)

	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.ErrorIs(err, store.ErrDuplicateFile)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathDoesntExist() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadata.State = ""

	collectionCountReturnsZero := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(0),
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

	subject := store.NewStore(&collectionCountReturnsZero, &suite.defaultkafkaProducer, suite.defaultClock)

	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.NoError(err)
}

func (suite *StoreSuite) TestRegisterFileUploadInsertReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	expectedError := errors.New("error occurred")

	collectionCountReturnsZero := mock.MongoCollectionMock{
		CountFunc:  CollectionCountReturnsValueAndNil(0),
		InsertFunc: CollectionInsertReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionCountReturnsZero, &suite.defaultkafkaProducer, suite.defaultClock)

	err := subject.RegisterFileUpload(suite.defaultContext, metadata)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenNotInCreatedState() {
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
			FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		}

		subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

		etagReference := files.FileEtagChange{
			Path: metadata.Path,
			Etag: metadata.Etag,
		}
		err := subject.MarkUploadComplete(suite.defaultContext, etagReference)

		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenFileNotExists() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StateCreated

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkUploadComplete(suite.defaultContext, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkUploadComplete(suite.defaultContext, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkUploadCompleteSucceeds() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateManyReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkUploadComplete(suite.defaultContext, etagReference)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenNotInCreatedState() {
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
			FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		}

		subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

		etagReference := files.FileEtagChange{
			Path: metadata.Path,
			Etag: metadata.Etag,
		}
		err := subject.MarkFileDecrypted(suite.defaultContext, etagReference)

		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenFileNotExists() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StatePublished

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkFileDecrypted(suite.defaultContext, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkFileDecrypted(suite.defaultContext, etagReference)

	suite.Error(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedSucceeds() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)

	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkFileDecrypted(suite.defaultContext, etagReference)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsErrNoDocumentFound() {
	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsUnspecifiedError() {
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedCollectionIDNil() {
	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

	err := subject.MarkFilePublished(suite.defaultContext, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrCollectionIDNotSet)
}

func (suite *StoreSuite) TestMarkFilePublishedStateUploaded() {
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
			FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		}

		subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

		err := subject.MarkFilePublished(suite.defaultContext, suite.path)

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
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.defaultkafkaProducer, suite.defaultClock)

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
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
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
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
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
