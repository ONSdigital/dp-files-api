package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	hcCli     *health.Client
	authToken string
}

// New creates a new instance of Client for the service
func New(filesAPIURL, authToken string) *Client {
	return &Client{
		hcCli:     health.NewClient(serviceName, filesAPIURL),
		authToken: authToken,
	}
}

// NewWithHealthClient creates a new instance of Client, reusing the URL and Clienter
// from the provided health check client
func NewWithHealthClient(hcCli *health.Client, authToken string) *Client {
	return &Client{
		hcCli:     health.NewClientWithClienter(serviceName, hcCli.URL, hcCli.Client),
		authToken: authToken,
	}
}

// Checker calls files api health endpoint and returns a check object to the caller
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

// readResponseBody reads the response body and returns it as bytes
func readResponseBody(resp *http.Response) ([]byte, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return b, nil
}

// handleErrorResponse handles non-OK status codes and returns an appropriate error
func handleErrorResponse(resp *http.Response) error {
	b, err := readResponseBody(resp)
	if err != nil {
		return err
	}

	var errorResp ErrorResponse
	if jsonErr := json.Unmarshal(b, &errorResp); jsonErr == nil && len(errorResp.Errors) > 0 {
		return fmt.Errorf("%s: %s", errorResp.Errors[0].Code, errorResp.Errors[0].Description)
	}

	bodyStr := string(b)
	if bodyStr != "" {
		return errors.New(bodyStr)
	}

	return fmt.Errorf("API returned status %d with no error message", resp.StatusCode)
}
