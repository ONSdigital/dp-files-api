package sdk

import (
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		apiError *APIError
		expected string
	}{
		{
			name: "APIError with multiple errors",
			apiError: &APIError{
				StatusCode: 400,
				Errors: &api.JSONErrors{
					Error: []api.JSONError{
						{Code: "ValidationError", Description: "Example1 required"},
						{Code: "ValidationError", Description: "Example2 is required"},
					},
				},
			},
			expected: "API error: status code 400\n  - code: ValidationError, description: Example1 required\n  - code: ValidationError, description: Example2 is required",
		},
		{
			name: "APIError with no errors",
			apiError: &APIError{
				StatusCode: 500,
				Errors:     nil,
			},
			expected: "API error: status code 500",
		},
		{
			name: "APIError with empty errors",
			apiError: &APIError{
				StatusCode: 404,
				Errors: &api.JSONErrors{
					Error: []api.JSONError{},
				},
			},
			expected: "API error: status code 404",
		},
	}

	for _, tt := range tests {
		Convey("Given an "+tt.name, t, func() {
			Convey("When Error is called", func() {
				result := tt.apiError.Error()

				Convey("Then the expected error string is returned", func() {
					So(result, ShouldEqual, tt.expected)
				})
			})
		})
	}
}
