package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/jwt"
	"github.com/ONSdigital/dp-files-api/files"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	permsdk "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/gorilla/mux"
)

type GetFileMetadata func(ctx context.Context, path string) (files.StoredRegisteredMetaData, error)

var errInvalidServiceToken = errors.New("invalid service token")

func HandleGetFileMetadata(getMetadata GetFileMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")
		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}

		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleGetFileMetadataWithPermissions(
	getMetadata GetFileMetadata,
	authMiddleware auth.Middleware,
	permissionsChecker auth.PermissionsChecker,
	zebedeeClient auth.ZebedeeClient,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		w.Header().Add("Content-Type", "application/json")
		metadata, err := getMetadata(req.Context(), vars["path"])
		if err != nil {
			handleError(w, err)
			return
		}
		if authMiddleware == nil || permissionsChecker == nil || zebedeeClient == nil {
			handleError(w, ErrAuthConfigUnavailable)
			return
		}

		authToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), permsdk.BearerPrefix)
		if authToken == "" {
			handleError(w, ErrMissingAuthToken)
			return
		}

		entityData, err := getEntityData(req.Context(), authToken, authMiddleware, zebedeeClient)
		if err != nil {
			if errors.Is(err, jwt.ErrPublickeysEmpty) {
				handleError(w, err)
				return
			}
			handleError(w, ErrInvalidAuthToken)
			return
		}
		if entityData == nil {
			handleError(w, ErrInvalidAuthToken)
			return
		}

		attributes := getPermissionAttributesFromRequest(metadata)
		hasPermission, err := permissionsChecker.HasPermission(req.Context(), *entityData, "static-files:read", attributes)
		if err != nil {
			handleError(w, err)
			return
		}
		if !hasPermission {
			handleError(w, ErrForbidden)
			return
		}
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func getEntityData(
	ctx context.Context,
	authToken string,
	authMiddleware auth.Middleware,
	zebedeeClient auth.ZebedeeClient,
) (*permsdk.EntityData, error) {
	entityData, err := authMiddleware.Parse(authToken)
	if err == nil {
		return entityData, nil
	}

	if errors.Is(err, jwt.ErrPublickeysEmpty) {
		return nil, err
	}

	identityResponse, idErr := zebedeeClient.CheckTokenIdentity(ctx, authToken)
	if idErr != nil || identityResponse == nil || identityResponse.Identifier == "" {
		return nil, errInvalidServiceToken
	}

	return &permsdk.EntityData{UserID: identityResponse.Identifier}, nil
}

func getPermissionAttributesFromRequest(metadata files.StoredRegisteredMetaData) map[string]string {
	attributes := map[string]string{}
	if metadata.ContentItem == nil {
		return attributes
	}
	if metadata.ContentItem.DatasetID == "" || metadata.ContentItem.Edition == "" {
		return attributes
	}
	attributes["dataset_edition"] = metadata.ContentItem.DatasetID + "/" + metadata.ContentItem.Edition
	return attributes
}
