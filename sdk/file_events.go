package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
	dpNetRequest "github.com/ONSdigital/dp-net/v3/request"
)

// CreateFileEvent creates a new file event in the audit log
func (c *Client) CreateFileEvent(ctx context.Context, event files.FileEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal file event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/file-events", c.hcCli.URL), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	dpNetRequest.AddServiceTokenHeader(req, c.authToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer closeResponseBody(ctx, resp)

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("invalid request: %w", handleErrorResponse(resp))
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorised: %w", handleErrorResponse(resp))
	case http.StatusForbidden:
		return fmt.Errorf("forbidden: %w", handleErrorResponse(resp))
	default:
		return handleErrorResponse(resp)
	}
}
