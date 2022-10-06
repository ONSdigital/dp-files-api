package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	mockFiles "github.com/ONSdigital/dp-files-api/files/mock"
	"github.com/ONSdigital/dp-files-api/health"
	hcMock "github.com/ONSdigital/dp-files-api/health/mock"
	"github.com/ONSdigital/dp-files-api/mongo"
	mongoMock "github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/service"
	"github.com/ONSdigital/dp-files-api/service/mock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	. "github.com/smartystreets/goconvey/convey"
)

var svc service.Service

func TestClose(t *testing.T) {
	Convey("Having a correctly initialised service in publishing mode", t, func() {
		hc := &hcMock.CheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
		}
		m := &mongoMock.ClientMock{
			CollectionFunc: func(s string) *mongodriver.Collection {
				return &mongodriver.Collection{}
			},
		}
		hs := &mockFiles.HTTPServerMock{ListenAndServeFunc: func() error { return nil }}

		km := &mock.OurProducerMock{}

		am := &authMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
		}

		serviceList := &mock.ServiceContainerMock{
			GetMongoDBFunc:        func() mongo.Client { return m },
			GetClockFunc:          func() clock.Clock { return nil },
			GetHTTPServerFunc:     func() files.HTTPServer { return hs },
			GetHealthCheckFunc:    func() health.Checker { return hc },
			GetKafkaProducerFunc:  func() kafka.IProducer { return km },
			GetAuthMiddlewareFunc: func() auth.Middleware { return am },
		}
		svcErrors := make(chan error, 1)

		ctx := context.Background()
		cfg, _ := config.Get()
		cfg.IsPublishing = true
		svc, _ := service.Run(ctx, serviceList, svcErrors, cfg, &mux.Router{})

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return nil }

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 4)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")
			assert.Equal(t, registerHealthChecks[1].Name, "Authorization Middleware")
			assert.Equal(t, registerHealthChecks[2].Name, "jwt keys state health check")
			assert.Equal(t, registerHealthChecks[3].Name, "Kafka Producer")
			assert.NoError(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return errors.New("shutdown broke") }

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 4)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")
			assert.Equal(t, registerHealthChecks[1].Name, "Authorization Middleware")
			assert.Equal(t, registerHealthChecks[2].Name, "jwt keys state health check")
			assert.Equal(t, registerHealthChecks[3].Name, "Kafka Producer")
			assert.Error(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If service times out while shutting down, the Close operation fails with the expected error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(2 * time.Second)
				return nil
			}

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 4)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")
			assert.Equal(t, registerHealthChecks[1].Name, "Authorization Middleware")
			assert.Equal(t, registerHealthChecks[2].Name, "jwt keys state health check")
			assert.Equal(t, registerHealthChecks[3].Name, "Kafka Producer")
			assert.Error(t, svc.Close(ctx, 100*time.Millisecond))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})
	})

	Convey("Having a correctly initialised service in web mode", t, func() {
		hc := &hcMock.CheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
		}
		m := &mongoMock.ClientMock{
			CollectionFunc: func(s string) *mongodriver.Collection {
				return &mongodriver.Collection{}
			},
		}
		hs := &mockFiles.HTTPServerMock{ListenAndServeFunc: func() error { return nil }}

		km := &mock.OurProducerMock{}

		am := &authMock.MiddlewareMock{}

		serviceList := &mock.ServiceContainerMock{
			GetMongoDBFunc:        func() mongo.Client { return m },
			GetClockFunc:          func() clock.Clock { return nil },
			GetHTTPServerFunc:     func() files.HTTPServer { return hs },
			GetHealthCheckFunc:    func() health.Checker { return hc },
			GetKafkaProducerFunc:  func() kafka.IProducer { return km },
			GetAuthMiddlewareFunc: func() auth.Middleware { return am },
		}
		svcErrors := make(chan error, 1)

		ctx := context.Background()
		cfg, _ := config.Get()
		cfg.IsPublishing = false
		svc, _ := service.Run(ctx, serviceList, svcErrors, cfg, &mux.Router{})

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return nil }

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 1)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")

			assert.NoError(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return errors.New("shutdown broke") }

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 1)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")

			assert.Error(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If service times out while shutting down, the Close operation fails with the expected error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(2 * time.Second)
				return nil
			}

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 1)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")

			assert.Error(t, svc.Close(ctx, 100*time.Millisecond))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})
	})
}
