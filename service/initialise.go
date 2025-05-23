package service

import (
	"context"
	"errors"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/aws"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dps3 "github.com/ONSdigital/dp-s3/v3"
	"github.com/ONSdigital/log.go/v2/log"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"github.com/ONSdigital/dp-files-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	cfg            *config.Config
	buildTime      string
	gitCommit      string
	version        string
	mongo          mongo.Client
	httpServer     files.HTTPServer
	healthChecker  health.Checker
	authMiddleware auth.Middleware
	kafkaProducer  kafka.IProducer
	s3Client       aws.S3Clienter
	router         *mux.Router
}

// NewServiceList creates a new service list of dependent services with the provided initialiser
func NewServiceList(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string, router *mux.Router) (*ExternalServiceList, error) {
	e := &ExternalServiceList{
		cfg:       cfg,
		buildTime: buildTime,
		gitCommit: gitCommit,
		version:   version,
		router:    router,
	}

	if err := e.setup(ctx); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *ExternalServiceList) setup(ctx context.Context) error {
	if err := e.createHealthCheck(); err != nil {
		return err
	}

	if err := e.createAuthMiddleware(); err != nil {
		return err
	}

	if err := e.createMongo(); err != nil {
		return err
	}

	e.createHTTPServer()
	if err := e.createKafkaProducer(); err != nil {
		return err
	}

	if err := e.createS3(ctx); err != nil {
		return err
	}

	return nil
}

func (e *ExternalServiceList) createS3(ctx context.Context) (err error) {
	if e.cfg.LocalstackHost != "" {
		awsConfig, err := awsConfig.LoadDefaultConfig(ctx,
			awsConfig.WithRegion(e.cfg.AwsRegion),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		)
		if err != nil {
			log.Error(ctx, "failed to create aws config", err)
			return err
		}

		e.s3Client = dps3.NewClientWithConfig(e.cfg.PrivateBucketName, awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = awssdk.String(e.cfg.LocalstackHost)
			o.UsePathStyle = true
		})
		return nil
	}

	s3Client, err := dps3.NewClient(ctx, e.cfg.AwsRegion, e.cfg.PrivateBucketName)
	if err != nil {
		return err
	}
	e.s3Client = s3Client
	return nil
}

func (e *ExternalServiceList) createAuthMiddleware() (err error) {
	e.authMiddleware, err = auth.NewFeatureFlaggedMiddleware(context.Background(), &e.cfg.AuthConfig, nil)

	return
}

func (e *ExternalServiceList) createKafkaProducer() error {
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       e.cfg.KafkaConfig.Addr,
		Topic:             e.cfg.KafkaConfig.StaticFilePublishedTopic,
		MinBrokersHealthy: &e.cfg.KafkaConfig.ProducerMinBrokersHealthy,
		KafkaVersion:      &e.cfg.KafkaConfig.Version,
		MaxMessageBytes:   &e.cfg.KafkaConfig.MaxBytes,
	}

	if e.cfg.KafkaConfig.SecProtocol != "" {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			e.cfg.KafkaConfig.SecCACerts,
			e.cfg.KafkaConfig.SecClientCert,
			e.cfg.KafkaConfig.SecClientKey,
			e.cfg.KafkaConfig.SecSkipVerify,
		)
	}

	ctx := context.Background()

	p, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		return err
	}

	if !e.cfg.IsPublishing {
		// In Web mode we do not want to produce kafka messages
		p.Close(ctx)
	}
	e.kafkaProducer = p

	return nil
}

func (e *ExternalServiceList) createHTTPServer() {
	s := dphttp.NewServer(e.cfg.BindAddr, e.router)
	s.HandleOSSignals = false
	e.httpServer = s
}

func (e *ExternalServiceList) createMongo() (err error) {
	e.mongo, err = mongo.New(e.cfg.MongoConfig)
	return
}

func (e *ExternalServiceList) createHealthCheck() error {
	versionInfo, err := healthcheck.NewVersionInfo(e.buildTime, e.gitCommit, e.version)
	if err != nil {
		return err
	}
	hc := healthcheck.New(versionInfo, e.cfg.HealthCheckCriticalTimeout, e.cfg.HealthCheckInterval)
	e.healthChecker = &hc
	return nil
}

func (e *ExternalServiceList) GetHTTPServer() files.HTTPServer {
	return e.httpServer
}

func (e *ExternalServiceList) GetHealthCheck() health.Checker {
	return e.healthChecker
}

func (e *ExternalServiceList) GetMongoDB() mongo.Client {
	return e.mongo
}

func (e *ExternalServiceList) GetClock() clock.Clock {
	return clock.SystemClock{}
}

func (e *ExternalServiceList) GetKafkaProducer() kafka.IProducer {
	return e.kafkaProducer
}

func (e *ExternalServiceList) GetAuthMiddleware() auth.Middleware {
	return e.authMiddleware
}

func (e *ExternalServiceList) GetS3Clienter() aws.S3Clienter {
	return e.s3Client
}

func (e *ExternalServiceList) Shutdown(ctx context.Context) error {
	shutdownErr := false
	e.healthChecker.Stop()

	if err := e.mongo.Close(ctx); err != nil {
		shutdownErr = true
		log.Error(ctx, "failed to shutdown mongo", err)
	}

	if err := e.httpServer.Shutdown(ctx); err != nil {
		shutdownErr = true
		log.Error(ctx, "failed to shutdown HTTP server", err)
	}

	if err := e.authMiddleware.Close(ctx); err != nil {
		shutdownErr = true
		log.Error(ctx, "failed to shutdown Authorization Middleware", err)
	}

	if shutdownErr {
		return errors.New("failures occurred during shutdown")
	}

	return nil
}
