package files_test

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/files/mock"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

type StoreSuite struct {
	suite.Suite
	collectionID  string
	path          string
	context       context.Context
	clock         steps.TestClock
	kafkaProducer kafkatest.IProducerMock
}

func (suite *StoreSuite) SetupTest() {
	suite.collectionID = "123456"
	suite.path = "test.txt"
	suite.context = context.Background()
	suite.clock = steps.TestClock{}
	suite.kafkaProducer = kafkatest.IProducerMock{}
}

func (suite *StoreSuite) TestGetFileMetadataError() {
	collection := suite.generateCollectionMockFindOneWithError()

	store := files.NewStore(&collection, &suite.kafkaProducer, suite.clock)
	_, err := store.GetFileMetadata(suite.context, suite.path)

	suite.Equal(files.ErrFileNotRegistered, err)
}

func (suite *StoreSuite) TestGetFileMetadataSuccess() {
	expectedMetadata := suite.generateMetadata(suite.collectionID)

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	collection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
	}

	store := files.NewStore(&collection, &suite.kafkaProducer, suite.clock)
	actualMetadata, _ := store.GetFileMetadata(suite.context, suite.path)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataSuccessSingleResult() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBSONBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionMockFindWithSingleResult(metadataBSONBytes)

	store := files.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata}
	actualMetadata, _ := store.GetFilesMetadata(suite.context, suite.collectionID)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoResult() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBSONBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionMockFindWithSingleResult(metadataBSONBytes)

	store := files.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, _ := store.GetFilesMetadata(suite.context, "INVALID_COLLECTION_ID")

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestRegisterFileUploadCountReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)

	AlwaysFindsExistingCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, errors.New("error occurred")
		},
	}

	store := files.NewStore(&AlwaysFindsExistingCollection, &suite.kafkaProducer, suite.clock)

	err := store.RegisterFileUpload(suite.context, metadata)

	suite.Error(err)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathAlreadyExists() {
	metadata := suite.generateMetadata(suite.collectionID)

	AlwaysFindsExistingCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 1, nil
		},
	}

	store := files.NewStore(&AlwaysFindsExistingCollection, &suite.kafkaProducer, suite.clock)

	err := store.RegisterFileUpload(suite.context, metadata)

	suite.ErrorIs(err, files.ErrDuplicateFile)
}

func (suite *StoreSuite) TestRegisterFileUploadWhenFilePathDoesntExist() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = ""

	collectionCountReturnsZero := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, nil
		},
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			actualMetadata := document.(files.StoredRegisteredMetaData)

			testCurrentTime := suite.clock.GetCurrentTime()

			suite.Equal(files.StateCreated, actualMetadata.State)
			suite.Equal(testCurrentTime, actualMetadata.CreatedAt)
			suite.Equal(testCurrentTime, actualMetadata.LastModified)

			suite.assertImmutableFieldsUnchanged(metadata, actualMetadata)

			return nil, nil
		},
	}

	store := files.NewStore(&collectionCountReturnsZero, &suite.kafkaProducer, suite.clock)

	err := store.RegisterFileUpload(suite.context, metadata)

	suite.NoError(err)
}

func (suite *StoreSuite) TestRegisterFileUploadInsertReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)
	collectionCountReturnsZero := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, nil
		},
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			return nil, errors.New("error occurred")
		},
	}

	store := files.NewStore(&collectionCountReturnsZero, &suite.kafkaProducer, suite.clock)

	err := store.RegisterFileUpload(suite.context, metadata)

	suite.Error(err)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenNotInCreatedState() {
	metadata := suite.generateMetadata(suite.collectionID)

	tests := []struct {
		currentState string
		expectedErr  error
	}{
		{files.StateUploaded, files.ErrFileNotInPublishedState},
		{files.StatePublished, files.ErrFileNotInPublishedState},
		{files.StateDecrypted, files.ErrFileNotInPublishedState},
	}

	for _, test := range tests {
		metadata.State = test.currentState

		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
				bson.Unmarshal(metadataBytes, result)
				return nil
			},
		}

		store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

		etagReference := files.FileEtagChange{
			Path: metadata.Path,
			Etag: metadata.Etag,
		}
		err := store.MarkUploadComplete(suite.context, etagReference)

		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenFileNotExists() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = files.StateCreated

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := store.MarkUploadComplete(suite.context, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkUploadCompleteFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = files.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return &mongodriver.CollectionUpdateResult{}, errors.New("an error occurred")
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := store.MarkUploadComplete(suite.context, etagReference)

	suite.Error(err)
}

func (suite *StoreSuite) TestMarkUploadCompleteSucceeds() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = files.StateCreated
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return &mongodriver.CollectionUpdateResult{}, nil
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := store.MarkUploadComplete(suite.context, etagReference)

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

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenNotInCreatedState() {
	metadata := suite.generateMetadata(suite.collectionID)

	tests := []struct {
		currentState string
		expectedErr  error
	}{
		{files.StateCreated, files.ErrFileNotInPublishedState},
		{files.StateUploaded, files.ErrFileNotInPublishedState},
		{files.StateDecrypted, files.ErrFileNotInPublishedState},
	}

	for _, test := range tests {
		metadata.State = test.currentState

		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
				bson.Unmarshal(metadataBytes, result)
				return nil
			},
		}

		store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

		etagReference := files.FileEtagChange{
			Path: metadata.Path,
			Etag: metadata.Etag,
		}
		err := store.MarkFileDecrypted(suite.context, etagReference)

		suite.Error(err)
		suite.ErrorIs(err, test.expectedErr, "the actual err was %v", err)
	}
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenFileNotExists() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = files.StatePublished

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := store.MarkFileDecrypted(suite.context, etagReference)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrFileNotRegistered, "the metadata looked for was %v", metadata)
}

func (suite *StoreSuite) TestMarkFileDecryptedFailsWhenUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = files.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return &mongodriver.CollectionUpdateResult{}, errors.New("an error occurred")
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := store.MarkFileDecrypted(suite.context, etagReference)

	suite.Error(err)
}

func (suite *StoreSuite) TestMarkFileDecryptedSucceeds() {
	metadata := suite.generateMetadata(suite.collectionID)

	metadata.State = files.StatePublished
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return &mongodriver.CollectionUpdateResult{}, nil
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	etagReference := files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
	err := store.MarkFileDecrypted(suite.context, etagReference)

	suite.NoError(err)
}

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

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsErrNoDocumentFound() {
	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.MarkFilePublished(suite.context, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrFileNotRegistered)
}

func (suite *StoreSuite) TestMarkFilePublishedFindReturnsUnspecifiedError() {
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return expectedError
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.MarkFilePublished(suite.context, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedCollectionIDNil() {
	metadata := suite.generateMetadata("")
	metadata.CollectionID = nil
	metadataBytes, _ := bson.Marshal(metadata)

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

	err := store.MarkFilePublished(suite.context, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, files.ErrCollectionIDNotSet)
}

func (suite *StoreSuite) TestMarkFilePublishedStateUploaded() {
	notUploadedStates := []string{
		files.StateDecrypted,
		files.StateCreated,
		files.StatePublished,
	}

	for _, state := range notUploadedStates {
		metadata := suite.generateMetadata(suite.collectionID)
		metadata.State = state
		metadataBytes, _ := bson.Marshal(metadata)

		collectionWithUploadedFile := mock.MongoCollectionMock{
			FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
				bson.Unmarshal(metadataBytes, result)
				return nil
			},
		}

		store := files.NewStore(&collectionWithUploadedFile, &suite.kafkaProducer, suite.clock)

		err := store.MarkFilePublished(suite.context, suite.path)

		suite.Error(err)
		suite.ErrorIs(err, files.ErrFileNotInUploadedState, "%s is not %s", state, files.StateUploaded)
	}
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = files.StateUploaded
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

	err := store.MarkFilePublished(suite.context, suite.path)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateKafkaReturnsError() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = files.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)
	expectedError := errors.New("an error occurred")

	collectionWithUploadedFile := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
		UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
			return nil, nil
		},
	}

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: func(schema *avro.Schema, event interface{}) error {
			return expectedError
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &kafkaMock, suite.clock)

	err := store.MarkFilePublished(suite.context, suite.path)

	numberOfTimesKafkaSendCalled := len(kafkaMock.SendCalls())

	suite.Equal(1, numberOfTimesKafkaSendCalled)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestMarkFilePublishedUpdateKafkaDoesNotReturnError() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadata.State = files.StateUploaded
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

	kafkaMock := kafkatest.IProducerMock{
		SendFunc: func(schema *avro.Schema, event interface{}) error {
			return nil
		},
	}

	store := files.NewStore(&collectionWithUploadedFile, &kafkaMock, suite.clock)

	err := store.MarkFilePublished(suite.context, suite.path)

	numberOfTimesKafkaSendCalled := len(kafkaMock.SendCalls())

	suite.Equal(1, numberOfTimesKafkaSendCalled)
	suite.NoError(err)
}

func (suite *StoreSuite) assertImmutableFieldsUnchanged(metadata, actualMetadata files.StoredRegisteredMetaData) {
	suite.Equal(metadata.Path, actualMetadata.Path)
	suite.Equal(metadata.IsPublishable, actualMetadata.IsPublishable)
	suite.Equal(metadata.CollectionID, actualMetadata.CollectionID)
	suite.Equal(metadata.Title, actualMetadata.Title)
	suite.Equal(metadata.SizeInBytes, actualMetadata.SizeInBytes)
	suite.Equal(metadata.Type, actualMetadata.Type)
	suite.Equal(metadata.Licence, actualMetadata.Licence)
	suite.Equal(metadata.LicenceUrl, actualMetadata.LicenceUrl)
	suite.Equal(metadata.Etag, actualMetadata.Etag)
}

func (suite *StoreSuite) generateCollectionMockFindOneWithError() mock.MongoCollectionMock {
	return mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}
}

func (suite *StoreSuite) generateCollectionMockFindWithSingleResult(metadataBytes []byte) mock.MongoCollectionMock {
	return mock.MongoCollectionMock{
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			result := files.StoredRegisteredMetaData{}
			bson.Unmarshal(metadataBytes, &result)

			resultPointer := results.(*[]files.StoredRegisteredMetaData)

			if filter.(primitive.M)["collection_id"] == *result.CollectionID {
				*resultPointer = []files.StoredRegisteredMetaData{result}
			}

			return 1, nil
		},
	}
}

func (suite *StoreSuite) generateTestTime(addedSeconds time.Duration) time.Time {
	return time.Now().Add(time.Second * addedSeconds).Round(time.Second).UTC()
}

func (suite *StoreSuite) generateMetadata(collectionID string) files.StoredRegisteredMetaData {
	createdAt := suite.generateTestTime(1)
	lastModified := suite.generateTestTime(2)
	uploadCompletedAt := suite.generateTestTime(3)
	publishedAt := suite.generateTestTime(4)
	decryptedAt := suite.generateTestTime(5)

	return files.StoredRegisteredMetaData{
		Path:              suite.path,
		IsPublishable:     false,
		CollectionID:      &collectionID,
		Title:             "Test file",
		SizeInBytes:       10,
		Type:              "text/plain",
		Licence:           "MIT",
		LicenceUrl:        "https://opensource.org/licenses/MIT",
		CreatedAt:         createdAt,
		LastModified:      lastModified,
		UploadCompletedAt: &uploadCompletedAt,
		PublishedAt:       &publishedAt,
		DecryptedAt:       &decryptedAt,
		State:             files.StateDecrypted,
		Etag:              "1234567",
	}
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}
