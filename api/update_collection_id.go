package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UpdateCollectionID func(ctx context.Context, path, collectionID string) error

type CollectionChange struct {
	CollectionID string `json:"collection_id"`
}

func HandlerUpdateCollectionID(updateCollectionID UpdateCollectionID) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cc := CollectionChange{}
		if err := json.NewDecoder(req.Body).Decode(&cc); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := updateCollectionID(req.Context(), mux.Vars(req)["path"], cc.CollectionID); err != nil {
			handleError(w, err)
		}
	}
}
