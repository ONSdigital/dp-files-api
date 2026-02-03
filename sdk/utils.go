package sdk

import (
	"context"
	"encoding/json"
	"io"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
)

// unmarshalJSONErrors unmarshals the JSON errors from the response body.
// If the body does not match the expected structure for JSON errors, it logs the error and returns nil.
func unmarshalJSONErrors(ctx context.Context, body io.ReadCloser) (*api.JSONErrors, error) {
	if body == nil {
		return nil, nil
	}

	var jsonErrors api.JSONErrors

	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &jsonErrors); err != nil {
		// Body did not match expected structure for JSON errors.
		// This case is only expected to occur when authorisation middleware returns an error.
		log.Error(ctx, "body did not match expected structure for JSON errors", err, log.Data{"body": string(bytes)})
		return nil, nil
	}

	return &jsonErrors, nil
}

// unmarshalStoredRegisteredMetaData unmarshals the StoredRegisteredMetaData from the response body
func unmarshalStoredRegisteredMetaData(body io.ReadCloser) (*files.StoredRegisteredMetaData, error) {
	if body == nil {
		return nil, ErrMissingResponseBody
	}

	var metadata files.StoredRegisteredMetaData

	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func stringToPointer(s string) *string {
	return &s
}
