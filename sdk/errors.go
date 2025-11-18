package sdk

import (
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-files-api/api"
)

// APIError represents an error returned by the files API
type APIError struct {
	StatusCode int
	Errors     *api.JSONErrors
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	errorMessage := fmt.Sprintf("API error: status code %d", e.StatusCode)

	if e.Errors == nil || len(e.Errors.Error) == 0 {
		return errorMessage
	}

	for _, err := range e.Errors.Error {
		errorMessage += fmt.Sprintf("\n  - code: %s, description: %s", err.Code, err.Description)
	}
	return errorMessage
}

// List of errors that can be returned by the SDK
var (
	ErrMissingResponseBody = errors.New("missing response body")
)
