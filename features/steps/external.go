package steps

import (
	"context"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

type External struct {
	Server      *dphttp.Server
	MongoClient mongo.Client
}

func (e *External) DoGetMongoDB(ctx context.Context, cfg *config.Config) (mongo.Client, error) {
	return e.MongoClient, nil
}

func (e *External) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (health.Checker, error) {
	hc := healthcheck.New(healthcheck.VersionInfo{}, time.Second, time.Second)
	return &hc, nil
}

func (e *External) DoGetHTTPServer(bindAddr string, r http.Handler) files.HTTPServer {
	e.Server.Server.Addr = bindAddr
	e.Server.Server.Handler = r

	return e.Server
}

func (e External) DoGetClock(ctx context.Context) clock.Clock {
	return testClock{}
}
