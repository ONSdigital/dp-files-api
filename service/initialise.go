package service

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
	"net/http"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	config        *config.Config
	buildTime     string
	gitCommit     string
	version       string
	mongo         mongo.Client
	httpServer    files.HTTPServer
	healthChecker health.Checker
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(cfg *config.Config, buildTime, gitCommit, version string) *ExternalServiceList {
	return &ExternalServiceList{
		config:    cfg,
		buildTime: buildTime,
		gitCommit: gitCommit,
		version:   version,
	}
}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(router http.Handler) files.HTTPServer {
	s := dphttp.NewServer(e.config.BindAddr, router)
	s.HandleOSSignals = false
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck() (health.Checker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(e.buildTime, e.gitCommit, e.version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, e.config.HealthCheckCriticalTimeout, e.config.HealthCheckInterval)
	return &hc, nil
}

func (e *ExternalServiceList) GetMongoDB(ctx context.Context) (mongo.Client, error) {
	return mongo.New(e.config.MongoConfig)
}

func (e *ExternalServiceList) GetClock(ctx context.Context) clock.Clock {
	return clock.SystemClock{}
}

// GetKafkaProducer returns a kafka producer with the provided config
func (e *ExternalServiceList) GetKafkaProducer(ctx context.Context) (kafka.IProducer, error) {
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       e.config.KafkaConfig.Addr,
		Topic:             e.config.KafkaConfig.StaticFilePublishedTopic,
		MinBrokersHealthy: &e.config.KafkaConfig.ProducerMinBrokersHealthy,
		KafkaVersion:      &e.config.KafkaConfig.Version,
		MaxMessageBytes:   &e.config.KafkaConfig.MaxBytes,
	}
	//if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
	//	pConfig.SecurityConfig = kafka.GetSecurityConfig(
	//		cfg.KafkaConfig.SecCACerts,
	//		cfg.KafkaConfig.SecClientCert,
	//		cfg.KafkaConfig.SecClientKey,
	//		cfg.KafkaConfig.SecSkipVerify,
	//	)
	//}
	return kafka.NewProducer(ctx, pConfig)
}

func (e *ExternalServiceList) Shutdown(ctx context.Context) error {
	shutdownErr := false
	e.healthChecker.Stop()

	err := e.mongo.Close(ctx)
	if err != nil {
		shutdownErr = true
		log.Error(ctx, "failed to shutdown mongo", err)
	}

	err = e.httpServer.Shutdown(ctx)
	if err != nil {
		shutdownErr = true
		log.Error(ctx, "failed to shutdown HTTP server", err)
	}

	if shutdownErr {
		return errors.New("failures occured durring shutdown")
	}

	return nil
}
