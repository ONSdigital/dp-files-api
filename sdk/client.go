package sdk

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	serviceName = "dp-files-api"
)

// Client is the SDK client for dp-files-api
type Client struct {
	hcCli *health.Client
}

// New creates a new instance of Client for the service
func New(filesAPIURL string) *Client {
	return &Client{
		hcCli: health.NewClient(serviceName, filesAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client, reusing the URL and Clienter
// from the provided health check client
func NewWithHealthClient(hcCli *health.Client) *Client {
	return &Client{
		hcCli: health.NewClientWithClienter(serviceName, hcCli.URL, hcCli.Client),
	}
}

// Checker Calls the health.Client's Checker method
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// Health returns the underlying Healthcheck Client for this API client
func (c *Client) Health() *health.Client {
	return c.hcCli
}

// URL returns the URL used by this client
func (c *Client) URL() string {
	return c.hcCli.URL
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
