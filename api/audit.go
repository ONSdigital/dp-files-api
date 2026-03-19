package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

func createAuditEvent(
	ctx context.Context,
	req *http.Request,
	authMiddleware auth.Middleware,
	idClient *clientsidentity.Client,
	createFileEvent CreateFileEvent,
	action string,
	resource string,
	fileMetadata *files.StoredRegisteredMetaData,
	logData log.Data,
) (int, error) {
	accessToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)
	if accessToken == "" {
		log.Info(ctx, "authorisation failed: no authorisation header in request", log.Classification(log.ProtectiveMonitoring), logData)
		return http.StatusUnauthorized, errors.New("no authorisation header in request")
	}

	authEntityData, err := getAuthEntityData(ctx, authMiddleware, idClient, accessToken, logData)
	if err != nil {
		log.Error(ctx, "failed to get auth entity data for audit record", err, logData)
		if strings.Contains(err.Error(), "key id unknown or invalid") || strings.Contains(err.Error(), "jwt token is malformed") {
			return http.StatusUnauthorized, err
		}
		return http.StatusForbidden, err
	}

	identityType := log.USER
	if authEntityData.IsServiceAuth {
		identityType = log.SERVICE
	}

	userID := authEntityData.EntityData.UserID
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
		return http.StatusInternalServerError, err
	}

	log.Info(ctx, "successfully created audit record", log.Classification(log.ProtectiveMonitoring), logAuthOption, logData)
	return http.StatusOK, nil
}
