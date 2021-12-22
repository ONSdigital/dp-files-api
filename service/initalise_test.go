package service

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files/mock"
	hcMock "github.com/ONSdigital/dp-files-api/health/mock"
	mongoMock "github.com/ONSdigital/dp-files-api/mongo/mock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServicesShutdownCalled(t *testing.T) {
	Convey("Shutting down dependencies in the service container", t, func() {

		m := &mongoMock.ClientMock{CloseFunc: func(ctx context.Context) error { return nil }}
		hc := &hcMock.CheckerMock{StopFunc: func() {}}
		hs := &mock.HTTPServerMock{ShutdownFunc: func(ctx context.Context) error { return nil }}

		serviceList := NewServiceList(&config.Config{}, "", "", "")
		serviceList.mongo = m
		serviceList.httpServer = hs
		serviceList.healthChecker = hc

		Convey("All dependencies successfully shutdown", func() {
			assert.NoError(t, serviceList.Shutdown(context.Background()))

			assert.Len(t, m.CloseCalls(), 1)
			assert.Len(t, hc.StopCalls(), 1)
			assert.Len(t, hs.ShutdownCalls(), 1)
		})

		Convey("Failure during one shutdown all other dependencies still shutdown", func() {
			m.CloseFunc = func(ctx context.Context) error { return errors.New("close error") }

			assert.Error(t, serviceList.Shutdown(context.Background()))

			assert.Len(t, m.CloseCalls(), 1)
			assert.Len(t, hc.StopCalls(), 1)
			assert.Len(t, hs.ShutdownCalls(), 1)
		})
	})
}
