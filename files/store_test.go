package files_test

import (
	"context"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/files/mock"
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
}

func (suite *StoreSuite) TestGetFileMetadataError() {
	collection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	ctx := context.Background()
	_, err := store.GetFileMetadata(ctx, "/data/test.txt")

	suite.Equal(files.ErrFileNotRegistered, err)
}

func (suite *StoreSuite) TestGetFileMetadataSuccess() {
	collectionID := "123456"
	expectedMetadata := suite.generateMetadata(collectionID)

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	collection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			bson.Unmarshal(metadataBytes, result)
			return nil
		},
	}

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	ctx := context.Background()
	actualMetadata, _ := store.GetFileMetadata(ctx, "/data/test.txt")

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataSuccessSingleResult() {
	collectionID := "123456"
	metadata := suite.generateMetadata(collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionWithSingleResultMock(metadataBytes)

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	ctx := context.Background()

	expectedMetadata := []files.StoredRegisteredMetaData{metadata}
	actualMetadata, _ := store.GetFilesMetadata(ctx, "123456")

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoResult() {
	collectionID := "123456"
	metadata := suite.generateMetadata(collectionID)
	metadataBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionWithSingleResultMock(metadataBytes)

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	ctx := context.Background()

	expectedMetadata := []files.StoredRegisteredMetaData{}
	actualMetadata, _ := store.GetFilesMetadata(ctx, "INVALID_COLLECTION_ID")

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) generateCollectionWithSingleResultMock(metadataBytes []byte) mock.MongoCollectionMock {
	return mock.MongoCollectionMock{
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			result := files.StoredRegisteredMetaData{}
			bson.Unmarshal(metadataBytes, &result)

			resultPointer := results.(*[]files.StoredRegisteredMetaData)

			bsonFilter := bson.M{"collection_id": result.CollectionID}
			if filter.(primitive.M)["collection_id"] == *(bsonFilter["collection_id"].(*string)) {
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
		Path:              "/data/test.txt",
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
