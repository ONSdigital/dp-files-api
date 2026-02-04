package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
)

func (c *Client) UpdateContentItem(ctx context.Context, filePath string, contentItem api.ContentItem, headers Headers) (*files.StoredRegisteredMetaData, error) {
	payload, err := json.Marshal(contentItem)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(c.hcCli.URL + "/files")
	if err != nil {
		return nil, err
	}

	// Remove leading slash so that JoinPath works if filePath starts with or without a "/"
	cleanedFilePath := strings.TrimPrefix(filePath, "/")
	url = url.JoinPath(cleanedFilePath)

	req, err := http.NewRequest(http.MethodPut, url.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	headers.Add(req)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		jsonErrors, err := unmarshalJSONErrors(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, &APIError{
			StatusCode: statusCode,
			Errors:     jsonErrors,
		}
	}

	metadata, err := unmarshalStoredRegisteredMetaData(resp.Body)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
