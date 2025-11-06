package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
)

type CreateFileEvent func(ctx context.Context, event *files.FileEvent) error

func HandlerCreateFileEvent(createFileEvent CreateFileEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var event files.FileEvent

		logData := log.Data{
			"method": req.Method,
			"path":   req.URL.Path,
		}

		if err := json.NewDecoder(req.Body).Decode(&event); err != nil {
			log.Error(req.Context(), "Failed to decode JSON request body", err, logData)
			writeError(w, buildGenericError("BadJson", "The JSON is not in a valid format"), http.StatusBadRequest)
			return
		}

		if err := validateFileEvent(&event); err != nil {
			writeError(w, buildGenericError("InvalidRequest", "Unable to process request due to a malformed or invalid request body or query parameter"), http.StatusBadRequest)
			log.Error(req.Context(), "File event validation failed", err, logData)
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

func validateFileEvent(event *files.FileEvent) error {
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

func buildGenericError(code, description string) jsonErrors {
	return jsonErrors{Error: []jsonError{{Code: code, Description: description}}}
}
