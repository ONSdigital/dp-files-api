package store_test

import (
	"context"
	"flag"
	"testing"

	"github.com/ONSdigital/dp-files-api/store"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	mongo "github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/stretchr/testify/suite"
	mongoRaw "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	componentFlag = flag.Bool("component", false, "perform component tests")
)

const (
	path = "testing.txt"
)

type StoreIntegrationTest struct {
	suite.Suite

	cfg   *config.Config
	mc    *mongo.Mongo
	ctx   context.Context
	store *store.Store
}

func (s *StoreIntegrationTest) SetupTest() {
	s.cfg, _ = config.Get()
	s.mc, _ = mongo.New(s.cfg.MongoConfig)
	s.ctx = context.Background()

	client, _ := mongoRaw.Connect(
		s.ctx,
		options.Client().ApplyURI("mongodb://root:password@mongo:27017"),
	)
	client.Database("files").Collection("metadata").Drop(s.ctx)

	cfg, _ := config.Get()
	s.store = store.NewStore(s.mc.Collection(config.MetadataCollection), s.mc.Collection(config.CollectionsCollection), s.mc.Collection(config.BundlesCollection), &kafkatest.IProducerMock{}, steps.TestClock{}, nil, cfg)
}

func TestStoreIntegration(t *testing.T) {
	if !*componentFlag {
		t.Skip("This test can only run in a docker environment")
	}

	suite.Run(t, new(StoreIntegrationTest))
}

func (s *StoreIntegrationTest) TestOptionalFieldsExcluded() {
	m := files.StoredRegisteredMetaData{
		Path:          path,
		IsPublishable: false,
		Title:         "Testing",
		SizeInBytes:   10,
		Type:          "text/plain",
		Licence:       "MIT",
		LicenceURL:    "www.licence.com/MIT",
	}

	s.store.RegisterFileUpload(s.ctx, m)

	out, _ := s.store.GetFileMetadata(s.ctx, path)

	s.Nil(out.UploadCompletedAt)
	s.Nil(out.PublishedAt)
	s.Nil(out.MovedAt)
	s.Nil(out.CollectionID)
}

func (s *StoreIntegrationTest) TestOptionalCollectionIDIncluded() {
	collectionID := "1234"

	m := files.StoredRegisteredMetaData{
		Path:          path,
		CollectionID:  &collectionID,
		IsPublishable: false,
		Title:         "Testing",
		SizeInBytes:   10,
		Type:          "text/plain",
		Licence:       "MIT",
		LicenceURL:    "www.licence.com/MIT",
	}

	s.store.RegisterFileUpload(s.ctx, m)

	out, _ := s.store.GetFileMetadata(s.ctx, path)

	s.Equal(collectionID, *out.CollectionID)
}
