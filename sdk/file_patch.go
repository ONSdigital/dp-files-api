package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/log.go/v2/log"
)

// FilePatchRequest represents the request payload for a PATCH request to "/files/{path:.*}"
// It includes StateMetadata and an optional ETag as some handlers require it
type FilePatchRequest struct {
	api.StateMetadata
	ETag string `json:"etag,omitempty"`
}

// patchFile sends a PATCH request to update the file metadata at the specified path
func (c *Client) patchFile(ctx context.Context, filePath string, patchReq FilePatchRequest, headers Headers) error {
	url, err := url.Parse(c.hcCli.URL + "/files")
	if err != nil {
		return err
	}

	// Remove leading slash so that JoinPath works if filePath starts with or without a "/"
	cleanedFilePath := strings.TrimPrefix(filePath, "/")
	url = url.JoinPath(cleanedFilePath)

	payload, err := json.Marshal(patchReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, url.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	headers.Add(req)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
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

// MarkFilePublished makes a PATCH request using patchFile to set the file state to "PUBLISHED"
func (c *Client) MarkFilePublished(ctx context.Context, filePath string, headers Headers) error {
	patchReq := FilePatchRequest{
		StateMetadata: api.StateMetadata{
			State: stringToPointer(store.StatePublished),
		},
	}
	return c.patchFile(ctx, filePath, patchReq, headers)
}

// MarkFileUploaded makes a PATCH request to set the file state to "UPLOADED" with an etag
func (c *Client) MarkFileUploaded(ctx context.Context, filePath, etag string, headers Headers) error {
	patchReq := FilePatchRequest{
		StateMetadata: api.StateMetadata{
			State: stringToPointer(store.StateUploaded),
		},
		ETag: etag,
	}
	return c.patchFile(ctx, filePath, patchReq, headers)
}
