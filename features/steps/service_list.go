package steps

import (
	"context"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

type fakeServiceContainer struct {
	server *dphttp.Server
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
	cfg, _ := config.Get()
	return mongo.New(cfg.MongoConfig)
}

func (e *fakeServiceContainer) GetClock(ctx context.Context) clock.Clock {
	return TestClock{}
}

func (e *fakeServiceContainer) GetKafkaProducer(ctx context.Context) (kafka.IProducer, error) {
	cfg, _ := config.Get()
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       cfg.KafkaConfig.Addr,
		Topic:             cfg.KafkaConfig.StaticFilePublishedTopic,
		MinBrokersHealthy: &cfg.KafkaConfig.ProducerMinBrokersHealthy,
		KafkaVersion:      &cfg.KafkaConfig.Version,
		MaxMessageBytes:   &cfg.KafkaConfig.MaxBytes,
	}

	return kafka.NewProducer(ctx, pConfig)
}

func (e *fakeServiceContainer) Shutdown(ctx context.Context) error {
	return nil
}
