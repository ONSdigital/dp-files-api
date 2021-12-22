package steps

import (
	"context"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

type fakeServiceContainer struct {
	server      *dphttp.Server
	mongoClient mongo.Client
}

func (e *fakeServiceContainer) GetHTTPServer(r http.Handler) files.HTTPServer {
	e.server.Server.Addr = ":26900"
	e.server.Server.Handler = r

	return e.server
}

func (e *fakeServiceContainer) GetHealthCheck() (health.Checker, error) {
	h := healthcheck.New(healthcheck.VersionInfo{}, time.Second, time.Second)
	return &h, nil
}

func (e *fakeServiceContainer) GetMongoDB(ctx context.Context) (mongo.Client, error) {
	return e.mongoClient, nil
}

func (e *fakeServiceContainer) GetClock(ctx context.Context) clock.Clock {
	return testClock{}
}

func (e *fakeServiceContainer) Shutdown(ctx context.Context) error {
	return nil
}
