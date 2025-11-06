package sdk

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// Error represents a single error in the error response
type Error struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
