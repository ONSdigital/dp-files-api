package files_test

import (
	"context"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/files/mock"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

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
