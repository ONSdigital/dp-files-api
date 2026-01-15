package sdk

import (
	"net/http"
)

const (
	Authorization string = "Authorization"
	BearerPrefix  string = "Bearer "
)

type Headers struct {
	Authorization string
}

// Add adds any provided headers to the request
func (h *Headers) Add(req *http.Request) {
	if h.Authorization != "" {
		req.Header.Set(Authorization, BearerPrefix+h.Authorization)
	}
}
