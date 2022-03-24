package store_test

import (
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestGetFileMetadataError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	subject := store.NewStore(&collection, &suite.defaultkafkaProducer, suite.defaultClock)
	_, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("file metadata not found", logEvent)
	suite.Equal(store.ErrFileNotRegistered, err)
}

func (suite *StoreSuite) TestGetFileMetadataSuccess() {
	expectedMetadata := suite.generateMetadata(suite.defaultCollectionID)

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultReturnsNil(metadataBytes),
	}

	subject := store.NewStore(&collection, &suite.defaultkafkaProducer, suite.defaultClock)
	actualMetadata, _ := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataSuccessSingleResult() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := mock.MongoCollectionMock{
		FindFunc: CollectionFindSetsResultAndReturns1IfCollectionIDMatchesFilter(metadataBytes),
	}

	subject := store.NewStore(&collection, &suite.defaultkafkaProducer, suite.defaultClock)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata}
	actualMetadata, _ := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoResult() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := mock.MongoCollectionMock{
		FindFunc: CollectionFindSetsResultAndReturns1IfCollectionIDMatchesFilter(metadataBytes),
	}

	subject := store.NewStore(&collection, &suite.defaultkafkaProducer, suite.defaultClock)

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, _ := subject.GetFilesMetadata(suite.defaultContext, "INVALID_COLLECTION_ID")

	suite.Exactly(expectedMetadata, actualMetadata)
}
