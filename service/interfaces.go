package service

import (
	"context"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"github.com/ONSdigital/dp-files-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"net/http"
)

//go:generate moq -out mock/serviceContainer.go -pkg mock . ServiceContainer
//go:generate moq -out mock/kafkaProducer.go -pkg mock . OurProducer

type OurProducer interface {
	kafka.IProducer
}

type ServiceContainer interface {
	GetHTTPServer(router http.Handler) files.HTTPServer
	GetHealthCheck() (health.Checker, error)
	GetMongoDB(ctx context.Context) (mongo.Client, error)
	GetClock(ctx context.Context) clock.Clock
	GetKafkaProducer(ctx context.Context) (kafka.IProducer, error)
	Shutdown(ctx context.Context) error
}

