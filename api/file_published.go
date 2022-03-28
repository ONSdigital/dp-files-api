package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type MarkFilePublished func(ctx context.Context, path string) error

func HandleMarkFilePublished(markFilePublished MarkFilePublished) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		path := mux.Vars(req)["path"]

		if err := markFilePublished(req.Context(), path); err != nil {
			handleError(w, err)
		}
		w.WriteHeader(http.StatusOK)
	}
}
