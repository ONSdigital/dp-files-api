package files_test

import (
	"context"
	"flag"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	mongo "github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	mongoRaw "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

var (
	componentFlag = flag.Bool("component", false, "perform component tests")
	loggingFlag   = flag.Bool("logging", false, "print logging")
)

func TestRegisterFileUpload(t *testing.T) {
	if !*componentFlag {
		t.Skip("This test can only run in a docker environment")
	}

	Convey("Ensure unset dates are left out of the document", t, func() {
		cfg, _ := config.Get()
		mc, _ := mongo.New(cfg.MongoConfig)
		ctx := context.Background()

		client, _ := mongoRaw.Connect(
			ctx,
			options.Client().ApplyURI("mongodb://root:password@mongo:27017"),
		)
		client.Database("files").Collection("metadata").Drop(ctx)

		s := files.NewStore(mc, &kafkatest.IProducerMock{}, steps.TestClock{})

		path := "testing.txt"
		Convey("Given a file has been registered", func() {

			m := files.StoredRegisteredMetaData{
				Path:          path,
				IsPublishable: false,
				CollectionID:  "1234567890",
				Title:         "Testing",
				SizeInBytes:   10,
				Type:          "text/plain",
				Licence:       "MIT",
				LicenceUrl:    "www.licence.com/MIT",
			}

			s.RegisterFileUpload(ctx, m)

			Convey("Then the UploadCompletedAt, PublishedAt & DecryptedAt fields do not exist in the document", func() {
				out, _ := s.GetFileMetadata(ctx, path)

				So(out.UploadCompletedAt, ShouldBeNil)
				So(out.PublishedAt, ShouldBeNil)
				So(out.DecryptedAt, ShouldBeNil)
			})
		})
	})
}
