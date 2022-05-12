package files

import "context"

//go:generate moq -out mock/server.go -pkg mock . HTTPServer

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
