package sdk_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-files-api/sdk"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filesAPIURL = "http://localhost:26900"
	authToken   = "test-auth-token"
	service     = "dp-files-api"
)

var ctx = context.Background()

func createHTTPClientMock(statusCode int) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(http.NoBody),
				Header:     http.Header{},
			}, nil
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func TestClient(t *testing.T) {
	Convey("Given a new files API client", t, func() {
		client := sdk.New(filesAPIURL, authToken)

		Convey("URL() method returns correct url", func() {
			So(client.URL(), ShouldEqual, filesAPIURL)
		})

		Convey("Health() method returns correct health client", func() {
			So(client.Health(), ShouldNotBeNil)
			So(client.Health().Name, ShouldEqual, service)
			So(client.Health().URL, ShouldEqual, filesAPIURL)
		})
	})
}

func TestNewWithHealthClient(t *testing.T) {
	Convey("Given a health check client that returns 200 OK", t, func() {
		mockHTTPClient := createHTTPClientMock(http.StatusOK)
		healthClient := health.NewClientWithClienter(service, filesAPIURL, mockHTTPClient)
		client := sdk.NewWithHealthClient(healthClient, authToken)
		initialStateCheck := health.CreateCheckState(service)

		Convey("URL() method returns correct url", func() {
			So(client.URL(), ShouldEqual, filesAPIURL)
		})

		Convey("Health() method returns correct health client", func() {
			So(client.Health(), ShouldNotBeNil)
			So(client.Health().Name, ShouldEqual, service)
			So(client.Health().URL, ShouldEqual, filesAPIURL)
		})

		Convey("Checker() method returns expected check", func() {
			err := client.Checker(ctx, &initialStateCheck)
			So(err, ShouldBeNil)
			So(initialStateCheck.Name(), ShouldEqual, service)
			So(initialStateCheck.Status(), ShouldEqual, healthcheck.StatusOK)
		})
	})
}
