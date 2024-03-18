package aws

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/aws/aws-sdk-go/service/s3"
)

//go:generate moq -out mock/s3.go -pkg mock_aws . S3Clienter

type S3Clienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Head(key string) (*s3.HeadObjectOutput, error)
}
