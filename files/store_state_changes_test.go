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
)

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
