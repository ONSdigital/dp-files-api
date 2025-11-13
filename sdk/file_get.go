package sdk

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-files-api/files"

	dpNetRequest "github.com/ONSdigital/dp-net/v3/request"
)

// GetFile retrieves the metadata for a file at the specified path
func (c *Client) GetFile(ctx context.Context, filePath string) (*files.StoredRegisteredMetaData, error) {
	url, err := url.Parse(c.hcCli.URL + "/files")
	if err != nil {
		return nil, err
	}

	// Remove leading slash so that JoinPath works if filePath starts with or without a "/"
	cleanedFilePath := strings.TrimPrefix(filePath, "/")
	url = url.JoinPath(cleanedFilePath)

	req, err := http.NewRequest(http.MethodGet, url.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	dpNetRequest.AddServiceTokenHeader(req, c.authToken)

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
