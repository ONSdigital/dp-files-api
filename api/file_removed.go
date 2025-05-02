package api

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type RemoveFile func(ctx context.Context, path string) error

func HandleRemoveFile(removeFile RemoveFile) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		path := mux.Vars(req)["path"]

		if err := removeFile(req.Context(), path); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
