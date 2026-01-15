package sdk

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filesAPIURL   = "http://localhost:26900"
	testAuthToken = "test-auth-token"
)

var (
	testHeaders = Headers{
		Authorization: testAuthToken,
	}
)

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

func newMockClienter(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(_ []string) {
		},
		DoFunc: func(_ context.Context, _ *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/health"}
		},
	}
}

func newMockFilesAPIClient(mockClienter *dphttp.ClienterMock) *Client {
	return NewWithHealthClient(health.NewClientWithClienter(serviceName, filesAPIURL, mockClienter))
}

func TestClient(t *testing.T) {
	Convey("Given a new files API client", t, func() {
		client := New(filesAPIURL)

		Convey("URL() method returns correct url", func() {
			So(client.URL(), ShouldEqual, filesAPIURL)
		})

		Convey("Health() method returns correct health client", func() {
			So(client.Health(), ShouldNotBeNil)
			So(client.Health().Name, ShouldEqual, serviceName)
			So(client.Health().URL, ShouldEqual, filesAPIURL)
		})
	})
}

func TestNewWithHealthClient(t *testing.T) {
	Convey("Given a health check client that returns 200 OK", t, func() {
		mockHTTPClient := createHTTPClientMock(http.StatusOK)
		healthClient := health.NewClientWithClienter(serviceName, filesAPIURL, mockHTTPClient)
		client := NewWithHealthClient(healthClient)
		initialStateCheck := health.CreateCheckState(serviceName)

		Convey("URL() method returns correct url", func() {
			So(client.URL(), ShouldEqual, filesAPIURL)
		})

		Convey("Health() method returns correct health client", func() {
			So(client.Health(), ShouldNotBeNil)
			So(client.Health().Name, ShouldEqual, serviceName)
			So(client.Health().URL, ShouldEqual, filesAPIURL)
		})

		Convey("Checker() method returns expected check", func() {
			err := client.Checker(context.Background(), &initialStateCheck)
			So(err, ShouldBeNil)
			So(initialStateCheck.Name(), ShouldEqual, serviceName)
			So(initialStateCheck.Status(), ShouldEqual, healthcheck.StatusOK)
		})
	})
}
