package store_test

import (
	"context"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/dp-files-api/store/mock"
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

type CollectionCountFunc func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error)
type CollectionFindFunc func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error)
type CollectionFindOneFunc func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error
type CollectionUpdateFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)
type CollectionUpdateManyFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)
type CollectionInsertFunc func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error)
type KafkaSendFunc func(schema *avro.Schema, event interface{}) error

func CollectionFindOneSetsResultReturnsNil(metadataBytes []byte) CollectionFindOneFunc {
	return func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
		bson.Unmarshal(metadataBytes, result)
		return nil
	}
}

func CollectionFindReturnsValueAndError(value int, expectedError error) CollectionFindFunc {
	return func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
		return value, expectedError
	}
}

func CollectionFindOneReturnsError(expectedError error) CollectionFindOneFunc {
	return func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
		return expectedError
	}
}

func CollectionUpdateReturnsNilAndNil() CollectionUpdateFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, nil
	}
}

func CollectionUpdateReturnsNilAndError(expectedError error) CollectionUpdateFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, expectedError
	}
}

func CollectionCountReturnsValueAndError(value int, expectedError error) CollectionCountFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
		return value, expectedError
	}
}

func CollectionCountReturnsValueAndNil(value int) CollectionCountFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
		return value, nil
	}
}

func CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout() CollectionCountFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
		bsonFilter := filter.(primitive.M)

		// Note: refactoring will also change this test
		if bsonFilter["$and"] == nil {
			// Count of all files in collection
			return 1, nil
		}

		// Second count of files not in uploaded state
		return 0, nil
	}
}

func CollectionUpdateManyReturnsNilAndError(expectedError error) CollectionUpdateManyFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, expectedError
	}
}

func CollectionUpdateManyReturnsNilAndNil() CollectionUpdateManyFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, nil
	}
}

func CollectionFindSetsResultsReturnsValueAndNil(metadataBytes []byte, value int) CollectionFindFunc {
	return func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
		result := files.StoredRegisteredMetaData{}
		bson.Unmarshal(metadataBytes, &result)

		resultPointer := results.(*[]files.StoredRegisteredMetaData)
		*resultPointer = []files.StoredRegisteredMetaData{result}

		return value, nil
	}
}

func CollectionInsertReturnsNilAndError(expectedError error) CollectionInsertFunc {
	return func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
		return nil, expectedError
	}
}

func KafkaSendReturnsError(expectedError error) KafkaSendFunc {
	return func(schema *avro.Schema, event interface{}) error {
		return expectedError
	}
}

func KafkaSendReturnsNil() KafkaSendFunc {
	return func(schema *avro.Schema, event interface{}) error {
		return nil
	}
}

func (suite *StoreSuite) SetupTest() {
	suite.collectionID = "123456"
	suite.path = "test.txt"
	suite.context = context.Background()
	suite.clock = steps.TestClock{}
	suite.kafkaProducer = kafkatest.IProducerMock{}
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
		State:             store.StateDecrypted,
		Etag:              "1234567",
	}
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}
