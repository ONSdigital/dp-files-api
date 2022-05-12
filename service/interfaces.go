package service

import (
	"context"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"

	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"github.com/ONSdigital/dp-files-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

//go:generate moq -out mock/serviceContainer.go -pkg mock . ServiceContainer
//go:generate moq -out mock/kafkaProducer.go -pkg mock . OurProducer

type OurProducer interface {
	kafka.IProducer
}

type ServiceContainer interface {
	GetHTTPServer() files.HTTPServer
	GetHealthCheck() health.Checker
	GetMongoDB() mongo.Client
	GetClock() clock.Clock
	GetKafkaProducer() kafka.IProducer
	GetAuthMiddleware() auth.Middleware
	Shutdown(ctx context.Context) error
}
