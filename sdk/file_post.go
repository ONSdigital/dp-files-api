package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
)

// RegisterFile makes a POST request to register new file metadata
func (c *Client) RegisterFile(ctx context.Context, metadata files.StoredRegisteredMetaData, headers Headers) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.hcCli.URL+"/files", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	headers.Add(req)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusCreated {
		jsonErrors, err := unmarshalJSONErrors(resp.Body)
		if err != nil {
			return err
		}
		return &APIError{
			StatusCode: statusCode,
			Errors:     jsonErrors,
		}
	}

	return nil
}
