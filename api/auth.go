package api

import (
	"context"
	"net/http"
	"strings"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	"github.com/ONSdigital/log.go/v2/log"

	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
)

type AuthEntityData struct {
	EntityData    *permissionsAPISDK.EntityData
	IsServiceAuth bool
}

// getAuthEntityData returns the EntityData associated with the provided access token
func getAuthEntityData(ctx context.Context, authMiddleware auth.Middleware, idClient *clientsidentity.Client, accessToken string, logData log.Data) (*AuthEntityData, error) {
	var entityData *permissionsAPISDK.EntityData

	var isServiceAuth bool

	var err error
	if strings.Contains(accessToken, ".") {
		// check JWT token
		entityData, err = authMiddleware.Parse(accessToken)
		if err != nil {
			log.Error(ctx, "authorisation failed: unable to parse jwt", err, log.Classification(log.ProtectiveMonitoring), log.Auth(log.USER, ""), logData)
			return nil, err
		}

		isServiceAuth = false
	} else {
		// check service id token is valid
		resp, err := idClient.CheckTokenIdentity(ctx, accessToken, clientsidentity.TokenTypeService)
		if err != nil {
			log.Error(ctx, "authorisation failed: service token issue", err, log.Classification(log.ProtectiveMonitoring), log.Auth(log.SERVICE, ""), logData)
			return nil, err
		} else {
			entityData = &permissionsAPISDK.EntityData{UserID: resp.Identifier}
			isServiceAuth = true
		}
	}

	authEntityData := AuthEntityData{
		EntityData:    entityData,
		IsServiceAuth: isServiceAuth,
	}

	return &authEntityData, nil
}

// checks the user permission within a function to determine access to pre-publish data
func checkUserPermission(r *http.Request, logData log.Data, permission string, attributes map[string]string, permissionsChecker auth.PermissionsChecker, entityData *permissionsAPISDK.EntityData) bool {
	var authorised bool

	hasPermission, err := permissionsChecker.HasPermission(r.Context(), *entityData, permission, attributes)
	if err != nil {
		log.Error(r.Context(), "permissions check errored", err, logData)
		return false
	}

	if hasPermission {
		authorised = true
	}

	logData["authenticated"] = authorised

	return authorised
}
