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

type FilesApiComponent struct {
	DpHttpServer *dphttp.Server
	svc          *service.Service
	svcList      service.ServiceContainer
	ApiFeature   *componenttest.APIFeature
	errChan      chan error
	mongoClient  *mongo.Client
	cg           *kafka.ConsumerGroup
	msgs         map[string]files.FilePublished
	isPublishing bool
	isAuthorised bool
}

func NewFilesApiComponent() *FilesApiComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	d := &FilesApiComponent{
		DpHttpServer: s,
		errChan:      make(chan error),
	}

	log.Namespace = "dp-files-api"

	d.isPublishing = true

	return d
}

func (d *FilesApiComponent) Initialiser() (http.Handler, error) {
	r := &mux.Router{}
	d.svcList = &fakeServiceContainer{d.DpHttpServer, r, d.isAuthorised}
	d.svc, _ = service.Run(context.Background(), d.svcList, d.errChan, d.isPublishing, r)

	return d.DpHttpServer.Handler, nil
}

func (d *FilesApiComponent) Reset() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, _ := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI("mongodb://root:password@mongo:27017"))

	d.mongoClient = client
	d.mongoClient.Database("files").Collection("metadata").Drop(ctx)
	d.mongoClient.Database("files").CreateCollection(ctx, "metadata")
	d.mongoClient.Database("files").Collection("metadata").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{"path", 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	d.isPublishing = true
}

func (d *FilesApiComponent) Close() error {
	if d.cg != nil {
		d.cg.Stop()
	}

	cfg, _ := config.Get()

	if d.svc != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		return d.svc.Close(ctx, cfg.GracefulShutdownTimeout)
	}
	return nil
}
