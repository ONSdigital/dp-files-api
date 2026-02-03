package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
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
			log.Error(ctx, "failed to parse error response from files-api", err, log.Data{
				"status_code": statusCode,
			})
		}
		return &APIError{
			StatusCode: statusCode,
			Errors:     jsonErrors,
		}
	}

	return nil
}
