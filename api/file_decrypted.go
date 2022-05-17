package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"

	"github.com/gorilla/mux"
)

type MarkDecryptionComplete func(ctx context.Context, change files.FileEtagChange) error

func HandleMarkFileDecrypted(markDecryptionComplete MarkDecryptionComplete) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m := EtagChange{}
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err := markDecryptionComplete(req.Context(), generateFileEtagChange(m, mux.Vars(req)["path"])); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
