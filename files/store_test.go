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
	collectionID string
	context      context.Context
}

func (suite *StoreSuite) SetupTest() {
	suite.collectionID = "123456"
	suite.context = context.Background()
}

func (suite *StoreSuite) TestGetFileMetadataError() {
	collection := suite.generateCollectionMockFindOneWithError()

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	_, err := store.GetFileMetadata(suite.context, "/data/test.txt")

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

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	actualMetadata, _ := store.GetFileMetadata(suite.context, "/data/test.txt")

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataSuccessSingleResult() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBSONBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionMockFindWithSingleResult(metadataBSONBytes)

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})

	expectedMetadata := []files.StoredRegisteredMetaData{metadata}
	actualMetadata, _ := store.GetFilesMetadata(suite.context, suite.collectionID)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoResult() {
	metadata := suite.generateMetadata(suite.collectionID)
	metadataBSONBytes, _ := bson.Marshal(metadata)

	collection := suite.generateCollectionMockFindWithSingleResult(metadataBSONBytes)

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, _ := store.GetFilesMetadata(suite.context, "INVALID_COLLECTION_ID")

	suite.Exactly(expectedMetadata, actualMetadata)
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
