package steps

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-files-api/service"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/v2/log"
)

type FilesApiComponent struct {
	DpHttpServer *dphttp.Server
	svc          *service.Service
	svcList      *service.ExternalServiceList
	ApiFeature   *componenttest.APIFeature
	Mongo        *componenttest.MongoFeature
	errChan      chan error
}

func NewFilesApiComponent(murl string) *FilesApiComponent {
	buf := bytes.NewBufferString("")
	log.SetDestination(buf, buf)

	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	d := &FilesApiComponent{
		DpHttpServer: s,
		errChan:      make(chan error),
	}

	fmt.Println("handler created in new", d.DpHttpServer.Server.Handler)

	log.Namespace = "dp-files-api"

	cfg, _ := config.Get()
	cfg.MongoConfig.URI = murl
	cfg.MongoConfig.Database = "files"
	cfg.MongoConfig.Collection = "metadata"
	cfg.MongoConfig.IsSSL = false
	cfg.ConnectionTimeout = 15 * time.Second
	mc, _ := mongo.New(cfg)

	d.svcList = service.NewServiceList(&External{
		Server:      d.DpHttpServer,
		MongoClient: mc,
	})

	return d
}

func (d *FilesApiComponent) Initialiser() (http.Handler, error) {
	cfg, _ := config.Get()
	d.svc, _ = service.Run(context.Background(), cfg, d.svcList, "1", "1", "1", d.errChan)
	return d.DpHttpServer.Handler, nil
}

func (d *FilesApiComponent) Reset() {
	d.Mongo.Reset()
}

func (d *FilesApiComponent) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := d.svc.Close(ctx)

	d.Mongo.Close()
	return err
}
