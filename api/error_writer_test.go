package api_test

import (
	"errors"
	"net/http"
)

type ErrorWriter struct {
	status int
}

func (e *ErrorWriter) Header() http.Header {
	return http.Header{}
}

func (e *ErrorWriter) Write(i []byte) (int, error) {
	return 0, errors.New("broken")
}

func (e *ErrorWriter) WriteHeader(statusCode int) {
	e.status = statusCode
}
