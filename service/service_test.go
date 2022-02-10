package service_test

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/clock"
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
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var svc service.Service

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service in publishing mode", t, func() {
		hc := &hcMock.CheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
		}
		m := &mongoMock.ClientMock{}
		hs := &mockFiles.HTTPServerMock{ListenAndServeFunc: func() error { return nil }}

		km := &mock.OurProducerMock{}

		serviceList := &mock.ServiceContainerMock{
			GetMongoDBFunc:       func(ctx context.Context) (mongo.Client, error) { return m, nil },
			GetClockFunc:         func(ctx context.Context) clock.Clock { return nil },
			GetHTTPServerFunc:    func(router http.Handler) files.HTTPServer { return hs },
			GetHealthCheckFunc:   func() (health.Checker, error) { return hc, nil },
			GetKafkaProducerFunc: func(ctx context.Context) (kafka.IProducer, error) { return km, nil },
		}
		svcErrors := make(chan error, 1)

		ctx := context.Background()
		svc, _ := service.Run(ctx, serviceList, svcErrors, true)

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return nil }

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 2)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")
			assert.Equal(t, registerHealthChecks[1].Name, "Kafka Producer")
			assert.NoError(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return errors.New("shutdown broke") }

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 2)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")
			assert.Equal(t, registerHealthChecks[1].Name, "Kafka Producer")
			assert.Error(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If service times out while shutting down, the Close operation fails with the expected error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(2 * time.Second)
				return nil
			}

			registerHealthChecks := hc.AddCheckCalls()

			assert.Len(t, registerHealthChecks, 2)
			assert.Equal(t, registerHealthChecks[0].Name, "Mongo DB")
			assert.Equal(t, registerHealthChecks[1].Name, "Kafka Producer")
			assert.Error(t, svc.Close(ctx, 100*time.Millisecond))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})
	})

	Convey("Having a correctly initialised service in web mode", t, func() {
		hc := &hcMock.CheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
		}
		m := &mongoMock.ClientMock{}
		hs := &mockFiles.HTTPServerMock{ListenAndServeFunc: func() error { return nil }}

		km := &mock.OurProducerMock{}

		serviceList := &mock.ServiceContainerMock{
			GetMongoDBFunc:       func(ctx context.Context) (mongo.Client, error) { return m, nil },
			GetClockFunc:         func(ctx context.Context) clock.Clock { return nil },
			GetHTTPServerFunc:    func(router http.Handler) files.HTTPServer { return hs },
			GetHealthCheckFunc:   func() (health.Checker, error) { return hc, nil },
			GetKafkaProducerFunc: func(ctx context.Context) (kafka.IProducer, error) { return km, nil },
		}
		svcErrors := make(chan error, 1)

		ctx := context.Background()
		svc, _ := service.Run(ctx, serviceList, svcErrors, false)

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
