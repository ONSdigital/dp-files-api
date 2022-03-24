package store_test

import (
	"errors"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestUpdateCollectionIDFindReturnsErrNoDocumentFound() {
	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestUpdateCollectionIDFindReturnsUnspecifiedError() {
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.UpdateCollectionID(suite.context, "", suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDCollectionIDAlreadySet() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrCollectionIDAlreadySet)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateReturnsError() {
	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := subject.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateSuccess() {
	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionContainsOneUploadedFileWithNoCollectionID := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
		UpdateFunc:  CollectionUpdateReturnsNilAndNil(),
	}

	subject := store.NewStore(&collectionContainsOneUploadedFileWithNoCollectionID, &suite.kafkaProducer, suite.clock)

	err := subject.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCountReturnsError() {
	expectedError := errors.New("an error occurred during files count")

	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndError(0, expectedError),
	}

	subject := store.NewStore(&collectionCountReturnsError, &suite.kafkaProducer, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCountReturnsZero() {
	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(0),
	}

	subject := store.NewStore(&collectionCountReturnsError, &suite.kafkaProducer, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrNoFilesInCollection)
}

func (suite *StoreSuite) TestMarkCollectionPublishedWhenFileExistsInStateOtherThanUploaded() {
	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: CollectionCountReturnsValueAndNil(1),
	}

	subject := store.NewStore(&collectionCountReturnsError, &suite.kafkaProducer, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, store.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkCollectionPublishedPersistenceFailure() {
	expectedError := errors.New("an error occurred")
	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndError(expectedError),
	}

	subject := store.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedFindUpdatedErrored() {
	expectedError := errors.New("an error occurred")

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindFunc:       CollectionFindReturnsValueAndError(0, expectedError),
	}

	subject := store.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedPersistenceSuccess() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindFunc:       CollectionFindSetsResultsReturnsValueAndNil(metadataBytes, 1),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsNil(),
	}

	subject := store.NewStore(&collection, &kafkaMock, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkCollectionPublishedKafkaErrorDoesNotFailOperation() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	kafkaError := errors.New("an error occurred with Kafka")

	collection := mock.MongoCollectionMock{
		CountFunc:      CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout(),
		UpdateManyFunc: CollectionUpdateManyReturnsNilAndNil(),
		FindFunc:       CollectionFindSetsResultsReturnsValueAndNil(metadataBytes, 1),
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: KafkaSendReturnsError(kafkaError),
	}

	subject := store.NewStore(&collection, &kafkaMock, suite.clock)

	err := subject.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Equal(1, len(kafkaMock.SendCalls()))
	suite.NoError(err)
}
