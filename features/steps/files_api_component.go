package steps

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-files-api/clock"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-files-api/service"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
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

type External struct {
	Server      *dphttp.Server
	MongoClient mongo.Client
}

func (e *External) DoGetMongoDB(ctx context.Context, cfg *config.Config) (mongo.Client, error) {
	return e.MongoClient, nil
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

func (e External) DoGetClock(ctx context.Context) clock.Clock {
	return testClock{}
}

type testClock struct{}

func (tt testClock) GetCurrentTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2021-10-19T09:30:30Z")
	return t
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
	mc, _ := mongo.New(context.Background(), cfg)

	d.svcList = service.NewServiceList(&External{
		Server:      d.DpHttpServer,
		MongoClient: mc,
	})

	return d
}

func (d *FilesApiComponent) Initialiser() (http.Handler, error) {
	cfg, _ := config.Get()
	ctx := context.Background()
	d.svc, _ = service.Run(ctx, cfg, d.svcList, "1", "1", "1", d.errChan)
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
