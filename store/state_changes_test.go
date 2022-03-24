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
	metadata := suite.generateMetadata(suite.collectionID)

	expectedError := errors.New("error occurred")

	AlwaysFindsExistingCollection := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndError(0, expectedError),
	}

	store := store.NewStore(&AlwaysFindsExistingCollection, &suite.kafkaProducer, suite.clock)

	err := store.RegisterFileUpload(suite.context, metadata)

	suite.Error(err)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathAlreadyExists() {
	metadata := suite.generateMetadata(suite.collectionID)

	AlwaysFindsExistingCollection := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(1),
	}

	subject := store.NewStore(&AlwaysFindsExistingCollection, &suite.kafkaProducer, suite.clock)

	err := subject.RegisterFileUpload(suite.context, metadata)

	suite.ErrorIs(err, store.ErrDuplicateFile)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathDoesntExist() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = ""

	collectionCountReturnsZero := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(0),
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			actualMetadata := document.(files.StoredRegisteredMetaData)

			testCurrentTime := suite.clock.GetCurrentTime()

			suite.Equal(store.StateCreated, actualMetadata.State)
			suite.Equal(testCurrentTime, actualMetadata.CreatedAt)
			suite.Equal(testCurrentTime, actualMetadata.LastModified)

			suite.assertImmutableFieldsUnchanged(metadata, actualMetadata)

			return nil, nil
		},
	}

	subject := store.NewStore(&collectionCountReturnsZero, &suite.kafkaProducer, suite.clock)

	err := subject.RegisterFileUpload(suite.context, metadata)

	suite.NoError(err)
}

func (suite *StoreSuite) TestRegisterFileUploadInsertReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)
	expectedError := errors.New("error occurred")

	collectionCountReturnsZero := mock.MongoCollectionMock{
		CountFunc:  CollectionCountReturnsValueAndNil(0),
		InsertFunc: CollectionInsertReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionCountReturnsZero, &suite.kafkaProducer, suite.clock)

	err := subject.RegisterFileUpload(suite.context, metadata)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenNotInCreatedState() {
	metadata := suite.generateMetadata(suite.collectionID)

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

		subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

		etagReference := files.FileEtagChange{
			Path: metadata.Path,
			Etag: metadata.Etag,
		}
		err := subject.MarkUploadComplete(suite.context, etagReference)

		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenFileNotExists() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = store.StateCreated

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkUploadComplete(suite.context, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = store.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkUploadComplete(suite.context, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkUploadCompleteSucceeds() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = store.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateManyReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkUploadComplete(suite.context, etagReference)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenNotInCreatedState() {
	metadata := suite.generateMetadata(suite.collectionID)

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

		subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

		etagReference := files.FileEtagChange{
			Path: metadata.Path,
			Etag: metadata.Etag,
		}
		err := subject.MarkFileDecrypted(suite.context, etagReference)

		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenFileNotExists() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = store.StatePublished

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkFileDecrypted(suite.context, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkFileDecrypted(suite.context, etagReference)

	suite.Error(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedSucceeds() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = store.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := subject.MarkFileDecrypted(suite.context, etagReference)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsErrNoDocumentFound() {
	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.MarkFilePublished(suite.context, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsUnspecifiedError() {
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.MarkFilePublished(suite.context, suite.path)

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

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.MarkFilePublished(suite.context, suite.path)

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
		metadata := suite.generateMetadata(suite.collectionID)
		metadata.State = state
		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		}

		subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

		err := subject.MarkFilePublished(suite.context, suite.path)

		suite.Error(err)
		suite.ErrorIs(err, store.ErrFileNotInUploadedState, "%s is not %s", state, store.StateUploaded)
	}
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.MarkFilePublished(suite.context, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateKafkaReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)
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

	subject := store.NewStore(&collectionWithUploadedFile, &kafkaMock, suite.clock)

	err := subject.MarkFilePublished(suite.context, suite.path)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateKafkaDoesNotReturnError() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsNil(),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &kafkaMock, suite.clock)

	err := subject.MarkFilePublished(suite.context, suite.path)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.NoError(err)
}
