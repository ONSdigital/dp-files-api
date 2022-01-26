package steps

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"time"

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
}

func NewFilesApiComponent() *FilesApiComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	d := &FilesApiComponent{
		DpHttpServer: s,
		errChan:      make(chan error),
	}

	log.Namespace = "dp-files-api"

	d.svcList = &fakeServiceContainer{s}

	return d
}

func (d *FilesApiComponent) Initialiser() (http.Handler, error) {
	d.svc , _ = service.Run(context.Background(), d.svcList, d.errChan)

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
}

func (d *FilesApiComponent) Close() error {
	cfg, _ := config.Get()

	if d.svc != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		return d.svc.Close(ctx, cfg.GracefulShutdownTimeout)
	}
	return nil
}
