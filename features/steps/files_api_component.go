package steps

import (
	"bytes"
	"context"
	"fmt"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-files-api/service"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/v2/log"
	"net/http"
	"os"
	"time"
)

type FilesApiComponent struct {
	DpHttpServer *dphttp.Server
	svc          *service.Service
	svcList      *service.ExternalServiceList
	ApiFeature   *componenttest.APIFeature
	Mongo        *componenttest.MongoFeature
	errChan      chan error
}

type External struct {
	Server *dphttp.Server
	MongoClient mongo.Client
}

func (e *External) DoGetMongoDB(ctx context.Context, cfg *config.Config) (mongo.Client, error) {
	return mongo.New(ctx, cfg)
}

func (e *External) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	hc := healthcheck.New(healthcheck.VersionInfo{}, time.Second, time.Second)
	return &hc, nil
}

func (e *External) DoGetHTTPServer(bindAddr string, r http.Handler) service.HTTPServer {
	e.Server.Server.Addr = bindAddr
	e.Server.Server.Handler = r

	return e.Server
}

func NewFilesApiComponent(mongoUrl string) *FilesApiComponent {
	buf := bytes.NewBufferString("")
	log.SetDestination(buf, buf)

	d := &FilesApiComponent{
		DpHttpServer: dphttp.NewServer("", http.NewServeMux()),
		errChan:      make(chan error),
	}

	fmt.Println("handler created in new", d.DpHttpServer.Server.Handler)

	os.Setenv("MONGODB_BIND_ADDR", mongoUrl)
	os.Setenv("MONGODB_DATABASE", "files")
	os.Setenv("MONGODB_COLLECTION", "metadata")
	os.Setenv("MONGODB_ENABLE_READ_CONCERN", "true")
	os.Setenv("MONGODB_ENABLE_WRITE_CONCERN", "true")
	os.Setenv("MONGODB_QUERY_TIMEOUT", "30")
	os.Setenv("MONGODB_CONNECT_TIMEOUT", "30")
	os.Setenv("MONGODB_IS_SSL", "true")

	log.Namespace = "dp-files-api"

	d.svcList = service.NewServiceList(&External{
		Server: d.DpHttpServer,
	})

	return d
}

func (d *FilesApiComponent) Initialiser() (http.Handler, error) {
	cfg, _ := config.Get()
	ctx := context.Background()
	cfg.MongoConfig.URI = d.Mongo.Server.URI()
	d.svc, _ = service.Run(ctx, cfg, d.svcList, "1", "1", "1", d.errChan)

	return d.DpHttpServer.Handler, nil
}

func (d *FilesApiComponent) Reset() {
	d.Mongo.Reset()
}

func (d *FilesApiComponent) Close() {
	d.Mongo.Close()
}
