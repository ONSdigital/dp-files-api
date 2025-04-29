package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UpdateBundleID func(ctx context.Context, path, bundleID string) error

type BundleChange struct {
	BundleID string `json:"bundle_id"`
}

func HandlerUpdateBundleID(updateBundleID UpdateBundleID) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		bc := BundleChange{}
		if err := json.NewDecoder(req.Body).Decode(&bc); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := updateBundleID(req.Context(), mux.Vars(req)["path"], bc.BundleID); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
