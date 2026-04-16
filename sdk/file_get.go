package sdk

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-files-api/files"
)

// GetFile retrieves the metadata for a file at the specified path
func (c *Client) GetFile(ctx context.Context, filePath string, headers Headers) (*files.StoredRegisteredMetaData, error) {
	log.Info(ctx, "Hello from the Files API SDK")
	parsedURL, err := url.Parse(c.hcCli.URL + "/files")
	if err != nil {
		return nil, err
	}

	// Remove leading slash so that JoinPath works if filePath starts with or without a "/"
	cleanedFilePath := strings.TrimPrefix(filePath, "/")
	parsedURL = parsedURL.JoinPath(cleanedFilePath)

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	headers.Add(req)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		jsonErrors, unmarshalErr := unmarshalJSONErrors(ctx, resp.Body)
		if unmarshalErr != nil {
			return nil, unmarshalErr
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
