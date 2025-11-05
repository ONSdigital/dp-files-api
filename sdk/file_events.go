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

// CreateFileEvent creates a new file event in the audit log and returns the created event
func (c *Client) CreateFileEvent(ctx context.Context, event files.FileEvent) (*files.FileEvent, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal file event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/file-events", c.hcCli.URL), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	dpNetRequest.AddServiceTokenHeader(req, c.authToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer closeResponseBody(ctx, resp)

	switch resp.StatusCode {
	case http.StatusCreated:
		var createdEvent files.FileEvent
		if err := json.NewDecoder(resp.Body).Decode(&createdEvent); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &createdEvent, nil
	case http.StatusBadRequest:
		return nil, fmt.Errorf("invalid request: %w", handleErrorResponse(resp))
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorised: %w", handleErrorResponse(resp))
	case http.StatusForbidden:
		return nil, fmt.Errorf("forbidden: %w", handleErrorResponse(resp))
	default:
		return nil, handleErrorResponse(resp)
	}
}
