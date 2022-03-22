package files_test

import (
	"context"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/files/mock"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func TestGetFileMetadataError(t *testing.T) {
	collection := mock.MongoCollectionMock{
		FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
			return mongodriver.ErrNoDocumentFound
		},
	}

	store := files.NewStore(&collection, &kafkatest.IProducerMock{}, steps.TestClock{})
	ctx := context.Background()
	_, err := store.GetFileMetadata(ctx, "/data/test.txt")

	assert.Equal(t, files.ErrFileNotRegistered, err)
}

func generateTestTime(addedSeconds time.Duration) time.Time {
	return time.Now().Add(time.Second * addedSeconds).Round(time.Second).UTC()
}

func TestGetFileMetadataSuccess(t *testing.T) {
	collectionID := "123456"
	createdAt := generateTestTime(1)
	lastModified := generateTestTime(2)
	uploadCompletedAt := generateTestTime(3)
	publishedAt := generateTestTime(4)
	decryptedAt := generateTestTime(5)

	expectedMetadata := files.StoredRegisteredMetaData{
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

	assert.Exactly(t, expectedMetadata, actualMetadata)

}
