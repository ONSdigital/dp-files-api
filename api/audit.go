package api

import (
	"context"

	"github.com/ONSdigital/dp-files-api/files"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

func createAuditEvent(
	ctx context.Context,
	createFileEvent CreateFileEvent,
	entityData *permissionsAPISDK.EntityData,
	isServiceAuth bool,
	action string,
	resource string,
	fileMetadata *files.StoredRegisteredMetaData,
	logData log.Data,
) error {
	identityType := log.USER
	if isServiceAuth {
		identityType = log.SERVICE
	}

	userID := entityData.UserID
	logAuthOption := log.Auth(identityType, userID)

	auditEvent := &files.FileEvent{
		RequestedBy: &files.RequestedBy{
			ID: userID,
		},
		Action:   action,
		Resource: resource,
		File:     fileMetadata,
	}

	if err := createFileEvent(ctx, auditEvent); err != nil {
		log.Error(ctx, "failed to create audit record", err, log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
		return err
	}

	log.Info(ctx, "successfully created audit record", log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
	return nil
}
