package files_test

import (
	"context"
	"flag"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	mongo "github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/stretchr/testify/suite"
	mongoRaw "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

var (
	componentFlag = flag.Bool("component", false, "perform component tests")
	loggingFlag   = flag.Bool("logging", false, "print logging")
)

type StoreIntegrationTest struct {
	suite.Suite
}

func TestStoreIntegration(t *testing.T) {
	if !*componentFlag {
		t.Skip("This test can only run in a docker environment")
	}

	suite.Run(t, new(StoreIntegrationTest))
}

func (s *StoreIntegrationTest) TestOptionalTimeFields() {
	cfg, _ := config.Get()
	mc, _ := mongo.New(cfg.MongoConfig)
	ctx := context.Background()

	client, _ := mongoRaw.Connect(
		ctx,
		options.Client().ApplyURI("mongodb://root:password@mongo:27017"),
	)
	client.Database("files").Collection("metadata").Drop(ctx)

	store := files.NewStore(mc, &kafkatest.IProducerMock{}, steps.TestClock{})

	path := "testing.txt"

	collectionID := "1234567890"
	m := files.StoredRegisteredMetaData{
		Path:          path,
		IsPublishable: false,
		CollectionID:  &collectionID,
		Title:         "Testing",
		SizeInBytes:   10,
		Type:          "text/plain",
		Licence:       "MIT",
		LicenceUrl:    "www.licence.com/MIT",
	}

	store.RegisterFileUpload(ctx, m)

	out, _ := store.GetFileMetadata(ctx, path)

	s.Nil(out.UploadCompletedAt)
	s.Nil(out.PublishedAt)
	s.Nil(out.DecryptedAt)
}
