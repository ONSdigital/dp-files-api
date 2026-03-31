package sdk

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// DeleteFile deletes a file at the specified filePath
func (c *Client) DeleteFile(ctx context.Context, filePath string, headers Headers) error {
	parsedURL, err := url.Parse(c.hcCli.URL + "/files")
	if err != nil {
		return err
	}

	// Remove leading slash so that JoinPath works if filePath starts with or without a "/"
	cleanedFilePath := strings.TrimPrefix(filePath, "/")
	parsedURL = parsedURL.JoinPath(cleanedFilePath)

	req, err := http.NewRequest(http.MethodDelete, parsedURL.String(), http.NoBody)
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
	if statusCode != http.StatusNoContent {
		jsonErrors, err := unmarshalJSONErrors(ctx, resp.Body)
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
