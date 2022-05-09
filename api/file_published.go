package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type MarkFilePublished func(ctx context.Context, path string) error

func HandleMarkFilePublished(markFilePublished MarkFilePublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := markFilePublished(req.Context(), mux.Vars(req)["path"]); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
