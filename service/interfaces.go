package service

import (
	"context"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"net/http"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/mongo"
)

//go:generate moq -out mock/serviceContainer.go -pkg mock . ServiceContainer

type ServiceContainer interface {
	GetHTTPServer(router http.Handler) files.HTTPServer
	GetHealthCheck() (health.Checker, error)
	GetMongoDB(ctx context.Context) (mongo.Client, error)
	GetClock(ctx context.Context) clock.Clock
	Shutdown(ctx context.Context) error
}

