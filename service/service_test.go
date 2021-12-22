package service_test

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/files"
	mock4 "github.com/ONSdigital/dp-files-api/files/mock"
	"github.com/ONSdigital/dp-files-api/health"
	mock2 "github.com/ONSdigital/dp-files-api/health/mock"
	"github.com/ONSdigital/dp-files-api/mongo"
	mock3 "github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/service"
	"github.com/ONSdigital/dp-files-api/service/mock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var svc service.Service

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func() {
		hc := &mock2.CheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
		}
		m := &mock3.ClientMock{}
		hs := &mock4.HTTPServerMock{ListenAndServeFunc: func() error { return nil }}

		serviceList := &mock.ServiceContainerMock{
			GetMongoDBFunc:     func(ctx context.Context) (mongo.Client, error) { return m, nil },
			GetClockFunc:       func(ctx context.Context) clock.Clock { return nil },
			GetHTTPServerFunc:  func(router http.Handler) files.HTTPServer { return hs },
			GetHealthCheckFunc: func() (health.Checker, error) { return hc, nil },
		}
		svcErrors := make(chan error, 1)

		ctx := context.Background()
		svc, _ := service.Run(ctx, serviceList, svcErrors)

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return nil }

			assert.NoError(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error { return errors.New("shutdown broke") }

			assert.Error(t, svc.Close(ctx, 2*time.Second))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})

		Convey("If service times out while shutting down, the Close operation fails with the expected error", func() {
			serviceList.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(2 * time.Second)
				return nil
			}

			assert.Error(t, svc.Close(ctx, 100*time.Millisecond))
			assert.Len(t, serviceList.ShutdownCalls(), 1)
		})
	})
}
