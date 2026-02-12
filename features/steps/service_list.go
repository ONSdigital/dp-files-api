package steps

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	permsdk "github.com/ONSdigital/dp-permissions-api/sdk"

	"github.com/ONSdigital/dp-files-api/aws"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	dphttp "github.com/ONSdigital/dp-net/v3/http"
	dps3 "github.com/ONSdigital/dp-s3/v3"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeServiceContainer struct {
	server    *dphttp.Server
	r         *mux.Router
	component *FilesAPIComponent
}

type testAuthMiddleware struct {
	component *FilesAPIComponent
}

func (m *testAuthMiddleware) Require(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := parseBearerToken(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !isValidToken(token) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !m.component.isAuthorised {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		handlerFunc(w, r)
	}
}

func (m *testAuthMiddleware) RequireWithAttributes(permission string, handlerFunc http.HandlerFunc, getAttributes auth.GetAttributesFromRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := parseBearerToken(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !isValidToken(token) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !m.component.isAuthorised {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		attributes, err := getAttributes(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if m.component.allowedDatasetEdition != "" {
			if attributes == nil || attributes["dataset_edition"] != m.component.allowedDatasetEdition {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		handlerFunc(w, r)
	}
}

func (m *testAuthMiddleware) Close(ctx context.Context) error {
	return nil
}

func (m *testAuthMiddleware) Parse(token string) (*permsdk.EntityData, error) {
	if !isValidToken(token) {
		return nil, errors.New("invalid token")
	}
	return &permsdk.EntityData{UserID: "user"}, nil
}

func (m *testAuthMiddleware) HealthCheck(ctx context.Context, state *healthcheck.CheckState) error {
	if state != nil {
		_ = state.Update(healthcheck.StatusOK, "ok", http.StatusOK)
	}
	return nil
}

func (m *testAuthMiddleware) IdentityHealthCheck(ctx context.Context, state *healthcheck.CheckState) error {
	if state != nil {
		_ = state.Update(healthcheck.StatusOK, "ok", http.StatusOK)
	}
	return nil
}

func (e *fakeServiceContainer) GetAuthMiddleware() auth.Middleware {
	return &testAuthMiddleware{component: e.component}
}

func parseBearerToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", false
	}
	return token, true
}

func isValidToken(token string) bool {
	return token == "test-valid-jwt-token" || token == "valid-service"
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

func (e *fakeServiceContainer) GetS3Clienter() aws.S3Clienter {
	cfg, _ := config.Get()

	awsConfig, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		fmt.Println("S3 ERROR: " + err.Error())
	}

	return dps3.NewClientWithConfig(cfg.PrivateBucketName, awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = awssdk.String("http://localstack:4566")
		o.UsePathStyle = true
	})
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
