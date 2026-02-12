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

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			handleError(w, err)
			return
		}
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		authToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		if authToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		entityData, err := getEntityData(req.Context(), authToken, authMiddleware, zebedeeClient)
		if err != nil {
			if err == jwt.ErrPublickeysEmpty {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err == errInvalidServiceToken {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if entityData == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		attributes := datasetEditionAttributes(metadata)
		hasPermission, err := permissionsChecker.HasPermission(req.Context(), *entityData, "static-files:read", attributes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !hasPermission {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			handleError(w, err)
			return
		}
	}
}

func getEntityData(
	ctx context.Context,
	authToken string,
	authMiddleware auth.Middleware,
	zebedeeClient auth.ZebedeeClient,
) (*permsdk.EntityData, error) {
	if strings.Contains(authToken, "service") {
		identityResponse, err := zebedeeClient.CheckTokenIdentity(ctx, authToken)
		if err != nil {
			return nil, errInvalidServiceToken
		}
		if identityResponse == nil || identityResponse.Identifier == "" {
			return nil, errInvalidServiceToken
		}
		return &permsdk.EntityData{UserID: identityResponse.Identifier}, nil
	}

	return authMiddleware.Parse(authToken)
}

func datasetEditionAttributes(metadata files.StoredRegisteredMetaData) map[string]string {
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
