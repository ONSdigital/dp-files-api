package store_test

import (
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestGetFileMetadataError() {
	collection := suite.generateCollectionMockFindOneWithError()

	subject := store.NewStore(&collection, &suite.kafkaProducer, suite.clock)
	_, err := subject.GetFileMetadata(suite.context, suite.path)

	suite.Equal(store.ErrFileNotRegistered, err)
}

func (suite *StoreSuite) TestGetFileMetadataSuccess() {
	expectedMetadata := suite.generateMetadata(suite.collectionID)

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
	}

	subject := store.NewStore(&collection, &suite.kafkaProducer, suite.clock)
	actualMetadata, _ := subject.GetFileMetadata(suite.context, suite.path)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataSuccessSingleResult() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBSONBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionMockFindWithSingleResult(metadataBSONBytes)

	subject := store.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata}
	actualMetadata, _ := subject.GetFilesMetadata(suite.context, suite.collectionID)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoResult() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBSONBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionMockFindWithSingleResult(metadataBSONBytes)

	subject := store.NewStore(&collection, &suite.kafkaProducer, suite.clock)

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, _ := subject.GetFilesMetadata(suite.context, "INVALID_COLLECTION_ID")

	suite.Exactly(expectedMetadata, actualMetadata)
}
