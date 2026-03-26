package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

type GetFileEvents func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*files.EventsList, error)

func HandlerGetFileEvents(getFileEvents GetFileEvents, createFileEvent CreateFileEvent, authMiddleware auth.Middleware, idClient *clientsidentity.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		logData := log.Data{
			"method": req.Method,
			"path":   req.URL.Path,
		}

		accessToken := strings.TrimPrefix(req.Header.Get(dprequest.AuthHeaderKey), dprequest.BearerPrefix)
		if accessToken == "" {
			log.Info(ctx, "authorisation failed: no authorisation header in request", log.Classification(log.ProtectiveMonitoring), logData)
			writeError(w, buildGenericError("Unauthorised", "The user is unauthorised"), http.StatusUnauthorized)
			return
		}

		authEntityData, err := getAuthEntityData(ctx, authMiddleware, idClient, accessToken, logData)
		if err != nil {
			log.Error(ctx, "failed to get auth entity data for file-events read", err, logData)
			if strings.Contains(err.Error(), "key id unknown or invalid") || strings.Contains(err.Error(), "jwt token is malformed") {
				writeError(w, buildGenericError("Unauthorised", "the request was not authorised"), http.StatusUnauthorized)
				return
			}
			writeError(w, buildGenericError("Forbidden", "the request was not authorised - check token and user's permissions"), http.StatusForbidden)
			return
		}

		limit, offset, err := parsePaginationParams(req)
		if err != nil {
			writeError(w, buildErrors(err, "InvalidRequest"), http.StatusBadRequest)
			return
		}

		path := req.URL.Query().Get("path")

		after, before, err := parseDateTimeParams(req)
		if err != nil {
			writeError(w, buildErrors(err, "InvalidRequest"), http.StatusBadRequest)
			return
		}

		eventsList, err := getFileEvents(ctx, limit, offset, path, after, before)
		if err != nil {
			handleError(w, err)
			return
		}

		if err := createAuditEvent(ctx, createFileEvent, authEntityData.EntityData, authEntityData.IsServiceAuth, files.ActionRead, req.URL.Path, nil, logData); err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(eventsList); err != nil {
			handleError(w, err)
			return
		}
	}
}

func parsePaginationParams(req *http.Request) (limit, offset int, err error) {
	limit = 20
	offset = 0

	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val < 0 || val > 1000 {
			return 0, 0, store.ErrInvalidPagination
		}
		limit = val
	}

	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			return 0, 0, store.ErrInvalidPagination
		}
		offset = val
	}

	return limit, offset, nil
}

func parseDateTimeParams(req *http.Request) (after, before *time.Time, err error) {
	if afterStr := req.URL.Query().Get("after"); afterStr != "" {
		t, err := time.Parse(time.RFC3339, afterStr)
		if err != nil {
			return nil, nil, store.ErrInvalidPagination
		}
		after = &t
	}

	if beforeStr := req.URL.Query().Get("before"); beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return nil, nil, store.ErrInvalidPagination
		}
		before = &t
	}

	return after, before, nil
}
