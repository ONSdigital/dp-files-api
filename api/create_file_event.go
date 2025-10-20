package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-files-api/sdk"
)

type CreateFileEvent func(ctx context.Context, event *sdk.FileEvent) error

func HandlerCreateFileEvent(createFileEvent CreateFileEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var event sdk.FileEvent

		if err := json.NewDecoder(req.Body).Decode(&event); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := validateFileEvent(&event); err != nil {
			writeError(w, buildErrors(err, "InvalidRequest"), http.StatusBadRequest)
			return
		}

		if err := createFileEvent(req.Context(), &event); err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(&event); err != nil {
			handleError(w, err)
			return
		}
	}
}

func validateFileEvent(event *sdk.FileEvent) error {
	if event.RequestedBy == nil || event.RequestedBy.ID == "" {
		return errors.New("requested_by.id is required")
	}
	if event.Action == "" {
		return errors.New("action is required")
	}
	if event.Resource == "" {
		return errors.New("resource is required")
	}
	if event.File == nil || event.File.Path == "" {
		return errors.New("file.path is required")
	}
	return nil
}
