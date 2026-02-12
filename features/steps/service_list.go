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
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	permsdk "github.com/ONSdigital/dp-permissions-api/sdk"

	"github.com/ONSdigital/dp-files-api/aws"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v3/request"

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

func (e *fakeServiceContainer) GetAuthMiddleware() auth.Middleware {
	return &authMock.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
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
				if !e.component.isAuthorised {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				handlerFunc(w, r)
			}
		},
		ParseFunc: func(token string) (*permsdk.EntityData, error) {
			if !isValidToken(token) {
				return nil, errors.New("invalid token")
			}
			return &permsdk.EntityData{UserID: "user"}, nil
		},
		HealthCheckFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			if state != nil {
				_ = state.Update(healthcheck.StatusOK, "ok", http.StatusOK)
			}
			return nil
		},
		IdentityHealthCheckFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			if state != nil {
				_ = state.Update(healthcheck.StatusOK, "ok", http.StatusOK)
			}
			return nil
		},
		CloseFunc: func(ctx context.Context) error {
			return nil
		},
	}
}

func (e *fakeServiceContainer) GetPermissionsChecker() auth.PermissionsChecker {
	return &authMock.PermissionsCheckerMock{
		HasPermissionFunc: func(ctx context.Context, entityData permsdk.EntityData, permission string, attributes map[string]string) (bool, error) {
			if !e.component.isAuthorised {
				return false, nil
			}
			if e.component.allowedDatasetEdition != "" {
				if attributes == nil || attributes["dataset_edition"] != e.component.allowedDatasetEdition {
					return false, nil
				}
			}
			return true, nil
		},
		CloseFunc: func(ctx context.Context) error {
			return nil
		},
	}
}

func (e *fakeServiceContainer) GetZebedeeClient() auth.ZebedeeClient {
	return &authMock.ZebedeeClientMock{
		CheckTokenIdentityFunc: func(ctx context.Context, token string) (*dprequest.IdentityResponse, error) {
			if token != "valid-service" {
				return nil, errors.New("invalid service token")
			}
			return &dprequest.IdentityResponse{Identifier: "service-user"}, nil
		},
	}
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
	// #nosec G101 -- test-only tokens used by component auth scenarios
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
