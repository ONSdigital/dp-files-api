package steps

import (
	"context"
	"github.com/ONSdigital/dp-files-api/mongo"

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
	Mongo        *componenttest.MongoFeature
	errChan      chan error
}

func NewFilesApiComponent(murl string) *FilesApiComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	d := &FilesApiComponent{
		DpHttpServer: s,
		errChan:      make(chan error),
	}

	log.Namespace = "dp-files-api"

	cfg, _ := config.Get()
	cfg.MongoConfig.URI = murl
	cfg.MongoConfig.Database = "files"
	cfg.MongoConfig.Collection = "metadata"
	cfg.MongoConfig.IsSSL = false
	cfg.ConnectionTimeout = 15 * time.Second

	m, _ := mongo.New(cfg)

	d.svcList = &fakeServiceContainer{
		server:      s,
		mongoClient: m,
	}

	return d
}

func (d *FilesApiComponent) Initialiser() (http.Handler, error) {
	d.svc, _ = service.Run(context.Background(), d.svcList, d.errChan)
	return d.DpHttpServer.Handler, nil
}

func (d *FilesApiComponent) Reset() {
	d.Mongo.Reset()
}

func (d *FilesApiComponent) Close() error {
	cfg, _ := config.Get()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := d.svc.Close(ctx, cfg.GracefulShutdownTimeout)

	d.Mongo.Close()
	return err
}
