package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-files-api/store"
	"github.com/gorilla/mux"
)

type MarkBundlePublished func(ctx context.Context, bundleID string) error

func HandleMarkBundlePublished(markBundlePublished MarkBundlePublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		bundleID := mux.Vars(req)["bundleID"]

		if err := markBundlePublished(req.Context(), bundleID); err != nil && err != store.ErrNoFilesInBundle {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
