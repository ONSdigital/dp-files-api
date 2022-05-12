package health

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"net/http"
)

//go:generate moq -out mock/healthCheck.go -pkg mock . Checker

// Checker defines the required methods from Healthcheck
type Checker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}