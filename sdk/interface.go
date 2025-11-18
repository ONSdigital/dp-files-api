package sdk

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *healthcheck.CheckState) error
	Health() *health.Client
	URL() string

	CreateFileEvent(ctx context.Context, event files.FileEvent) (*files.FileEvent, error)
	DeleteFile(ctx context.Context, filePath string) error
	GetFile(ctx context.Context, filePath string) (*files.StoredRegisteredMetaData, error)
	MarkFilePublished(ctx context.Context, filePath string) error
}
