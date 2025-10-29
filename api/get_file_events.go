package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-files-api/sdk"
)

type GetFileEvents func(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*sdk.EventsList, error)

func HandlerGetFileEvents(getFileEvents GetFileEvents) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		eventsList, err := getFileEvents(req.Context(), limit, offset, path, after, before)
		if err != nil {
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
			return 0, 0, errors.New("unable to process request due to a malformed or invalid request body or query parameter")
		}
		limit = val
	}

	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			return 0, 0, errors.New("unable to process request due to a malformed or invalid request body or query parameter")
		}
		offset = val
	}

	return limit, offset, nil
}

func parseDateTimeParams(req *http.Request) (after, before *time.Time, err error) {
	if afterStr := req.URL.Query().Get("after"); afterStr != "" {
		t, err := time.Parse(time.RFC3339, afterStr)
		if err != nil {
			return nil, nil, errors.New("unable to process request due to a malformed or invalid request body or query parameter")
		}
		after = &t
	}

	if beforeStr := req.URL.Query().Get("before"); beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return nil, nil, errors.New("unable to process request due to a malformed or invalid request body or query parameter")
		}
		before = &t
	}

	return after, before, nil
}
