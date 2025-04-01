package steps

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/ONSdigital/dp-files-api/files"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/service"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/v2/log"
)

type FilesAPIComponent struct {
	DpHTTPServer *dphttp.Server
	svc          *service.Service
	svcList      service.ServiceContainer
	APIFeature   *componenttest.APIFeature
	errChan      chan error
	mongoClient  *mongo.Client
	cg           *kafka.ConsumerGroup
	msgs         map[string]files.FilePublished
	isPublishing bool
	isAuthorised bool
}

func NewFilesAPIComponent() *FilesAPIComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	c := &FilesAPIComponent{
		DpHTTPServer: s,
		errChan:      make(chan error),
	}

	log.Namespace = "dp-files-api"

	c.isPublishing = true

	return c
}

func (c *FilesAPIComponent) Initialiser() (http.Handler, error) {
	r := &mux.Router{}
	c.svcList = &fakeServiceContainer{c.DpHTTPServer, r, c.isAuthorised}
	cfg, _ := config.Get()
	cfg.IsPublishing = c.isPublishing
	c.svc, _ = service.Run(context.Background(), c.svcList, c.errChan, cfg, r)

	return c.DpHTTPServer.Handler, nil
}

func (c *FilesAPIComponent) Reset() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, _ := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI("mongodb://root:password@mongo:27017"))

	c.mongoClient = client
	if err := c.mongoClient.Database("files").Collection("metadata").Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop metadata collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").CreateCollection(ctx, "metadata"); err != nil {
		log.Error(ctx, "failed to create metadata collection", err)
		panic(err)
	}
	if _, err := c.mongoClient.Database("files").Collection("metadata").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "path", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on metadata collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").Collection("collections").Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop collections collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").CreateCollection(ctx, "collections"); err != nil {
		log.Error(ctx, "failed to create collections collection", err)
		panic(err)
	}
	if _, err := c.mongoClient.Database("files").Collection("collections").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on collections collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").Collection("bundles").Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop bundles collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").CreateCollection(ctx, "bundles"); err != nil {
		log.Error(ctx, "failed to create bundles collection", err)
		panic(err)
	}
	if _, err := c.mongoClient.Database("files").Collection("bundles").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on bundles collection", err)
		panic(err)
	}
	c.isPublishing = true
}

func (c *FilesAPIComponent) Close() error {
	if c.cg != nil {
		if err := c.cg.Stop(); err != nil {
			return err
		}
	}

	cfg, _ := config.Get()

	if c.svc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return c.svc.Close(ctx, cfg.GracefulShutdownTimeout)
	}
	return nil
}
