package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-files-api/store"
	"github.com/gorilla/mux"
)

type MarkCollectionPublished func(ctx context.Context, collectionID string) error

func HandleMarkCollectionPublished(markCollectionPublished MarkCollectionPublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		collectionID := mux.Vars(req)["collectionID"]

		if err := markCollectionPublished(req.Context(), collectionID); err != nil && err != store.ErrNoFilesInCollection {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
