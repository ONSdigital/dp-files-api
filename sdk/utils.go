package sdk

import (
	"encoding/json"
	"io"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
)

// unmarshalJsonErrors unmarshals the JSON errors from the response body.
// This function assumes the response body JSON structure matches api.JsonErrors
func unmarshalJsonErrors(body io.ReadCloser) (*api.JsonErrors, error) {
	if body == nil {
		return nil, nil
	}

	var jsonErrors api.JsonErrors

	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &jsonErrors); err != nil {
		return nil, err
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
