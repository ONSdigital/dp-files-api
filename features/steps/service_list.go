package steps

import (
	"context"
	"github.com/gorilla/mux"
	"time"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

type fakeServiceContainer struct {
	server *dphttp.Server
	r      *mux.Router
}

func (e *fakeServiceContainer) GetAuthMiddleware() auth.Middleware {
	return &authMock.MiddlewareMock{
		HealthCheckFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		CloseFunc:       func(ctx context.Context) error { return nil },
	}
}

func (e *fakeServiceContainer) GetHTTPServer() files.HTTPServer {
	e.server.Server.Addr = ":26900"
	e.server.Server.Handler = e.r

	return e.server
}

func (e *fakeServiceContainer) GetHealthCheck() health.Checker {
	h := healthcheck.New(healthcheck.VersionInfo{}, time.Second, time.Second)
	return &h
}

func (e *fakeServiceContainer) GetMongoDB() mongo.Client {
	cfg, _ := config.Get()
	m, _ := mongo.New(cfg.MongoConfig)
	return m
}

func (e *fakeServiceContainer) GetClock() clock.Clock {
	return TestClock{}
}

func (e *fakeServiceContainer) GetKafkaProducer() kafka.IProducer {
	cfg, _ := config.Get()
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       cfg.KafkaConfig.Addr,
		Topic:             cfg.KafkaConfig.StaticFilePublishedTopic,
		MinBrokersHealthy: &cfg.KafkaConfig.ProducerMinBrokersHealthy,
		KafkaVersion:      &cfg.KafkaConfig.Version,
		MaxMessageBytes:   &cfg.KafkaConfig.MaxBytes,
	}

	producer, _ := kafka.NewProducer(context.Background(), pConfig)
	return producer
}

func (e *fakeServiceContainer) Shutdown(ctx context.Context) error {
	return nil
}
