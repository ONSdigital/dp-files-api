package files_test

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/files/mock"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (suite *StoreSuite) TestUpdateCollectionIDFindReturnsErrNoDocumentFound() {
	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestUpdateCollectionIDFindReturnsUnspecifiedError() {
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return expectedError
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.UpdateCollectionID(suite.context, "", suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDCollectionIDAlreadySet() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrCollectionIDAlreadySet)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateReturnsError() {
	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, expectedError
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestUpdateCollectionIDUpdateSuccess() {
	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.UpdateCollectionID(suite.context, suite.path, suite.collectionID)

	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCountReturnsError() {
	ExpectedError := errors.New("an error occurred during files count")

	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, ExpectedError
		},
	}

	store := files.NewStore(&collectionCountReturnsError, &suite.kafkaProducer, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, ExpectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedCountReturnsZero() {
	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, nil
		},
	}

	store := files.NewStore(&collectionCountReturnsError, &suite.kafkaProducer, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrNoFilesInCollection)
}

func (suite *StoreSuite) TestMarkCollectionPublishedWhenFileExistsInStateOtherThanUploaded() {
	collectionCountReturnsError := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 1, nil
		},
	}

	store := files.NewStore(&collectionCountReturnsError, &suite.kafkaProducer, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrFileNotInUploadedState)
}

func (suite *StoreSuite) TestMarkCollectionPublishedPersistenceFailure() {
	expectedError := errors.New("an error occurred")
	collection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			bsonFilter := filter.(primitive.M)

			// Note: refactoring will also change this test
			if bsonFilter["$and"] == nil {
				// Count of all files in collection
				return 1, nil
			}

			// Second count of files not in uploaded state
			return 0, nil
		},
		UpdateManyFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return &mongodriver.CollectionUpdateResult{}, expectedError
		},
	}

	store := files.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedFindUpdatedErrored() {
	expectedError := errors.New("an error occurred")

	collection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			bsonFilter := filter.(primitive.M)

			// Note: refactoring will also change this test
			if bsonFilter["$and"] == nil {
				// Count of all files in collection
				return 1, nil
			}

			// Second count of files not in uploaded state
			return 0, nil
		},
		UpdateManyFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, expectedError
		},
	}

	store := files.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkCollectionPublishedPersistenceSuccess() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			bsonFilter := filter.(primitive.M)

			// Note: refactoring will also change this test
			if bsonFilter["$and"] == nil {
				// Count of all files in collection
				return 1, nil
			}

			// Second count of files not in uploaded state
			return 0, nil
		},
		UpdateManyFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			result := files.StoredRegisteredMetaData{}
			bson.Unmarshal(metadataBytes, &result)

			resultPointer := results.(*[]files.StoredRegisteredMetaData)
			*resultPointer = []files.StoredRegisteredMetaData{result}

			return 1, nil
		},
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: func(schema *avro.Schema, event interface{}) error {
			return nil
		},
	}

	store := files.NewStore(&collection, &kafkaMock, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	numberOfTimesKafkaCalled := len(kafkaMock.SendCalls())
	suite.Equal(1, numberOfTimesKafkaCalled)
	suite.NoError(err)
}

func (suite *StoreSuite) TestMarkCollectionPublishedKafkaErrorDoesNotFailOperation() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			bsonFilter := filter.(primitive.M)

			// Note: refactoring will also change this test
			if bsonFilter["$and"] == nil {
				// Count of all files in collection
				return 1, nil
			}

			// Second count of files not in uploaded state
			return 0, nil
		},
		UpdateManyFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			result := files.StoredRegisteredMetaData{}
			bson.Unmarshal(metadataBytes, &result)

			resultPointer := results.(*[]files.StoredRegisteredMetaData)
			*resultPointer = []files.StoredRegisteredMetaData{result}

			return 1, nil
		},
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: func(schema *avro.Schema, event interface{}) error {
			return errors.New("an error occurred with Kafka")
		},
	}

	store := files.NewStore(&collection, &kafkaMock, suite.clock)

	err := store.MarkCollectionPublished(suite.context, suite.collectionID)

	numberOfTimesKafkaCalled := len(kafkaMock.SendCalls())
	suite.Equal(1, numberOfTimesKafkaCalled)
	suite.NoError(err)
}
