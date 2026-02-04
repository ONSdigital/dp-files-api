package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
)

// CreateFileEvent creates a new file event in the audit log and returns the created event
func (c *Client) CreateFileEvent(ctx context.Context, event files.FileEvent, headers Headers) (*files.FileEvent, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal file event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/file-events", c.hcCli.URL), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	headers.Add(req)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer closeResponseBody(ctx, resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusCreated {
		jsonErrors, err := unmarshalJSONErrors(ctx, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, &APIError{
			StatusCode: statusCode,
			Errors:     jsonErrors,
		}
	}

	var createdEvent files.FileEvent
	if err := json.NewDecoder(resp.Body).Decode(&createdEvent); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &createdEvent, nil
}
