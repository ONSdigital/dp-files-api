package sdk

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *healthcheck.CheckState) error
	Health() *health.Client
	URL() string

	CreateFileEvent(ctx context.Context, event files.FileEvent, headers Headers) (*files.FileEvent, error)
	DeleteFile(ctx context.Context, filePath string, headers Headers) error
	GetFile(ctx context.Context, filePath string, headers Headers) (*files.StoredRegisteredMetaData, error)
	MarkFilePublished(ctx context.Context, filePath string, headers Headers) error
	RegisterFile(ctx context.Context, metadata files.StoredRegisteredMetaData, headers Headers) error
	MarkFileUploaded(ctx context.Context, filePath string, etag string, headers Headers) error
	ContentItemUpdate(ctx context.Context, filePath string, item api.ContentItem, headers Headers) (files.StoredRegisteredMetaData, error)
}
