package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/gorilla/mux"
)

type UpdateContentItem func(ctx context.Context, path string, contentItem files.StoredContentItem) (files.StoredRegisteredMetaData, error)

type ContentItemChange struct {
	ContentItem *ContentItem `json:"content_item"`
}

func HandlerUpdateContentItem(updateContentItem UpdateContentItem) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var cic = ContentItemChange{}
		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&cic); err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		storedContentItemChange := files.StoredContentItem{
			DatasetID: cic.ContentItem.DatasetID,
			Edition:   cic.ContentItem.Edition,
			Version:   cic.ContentItem.Version,
		}

		metadata, err := updateContentItem(req.Context(), mux.Vars(req)["path"], storedContentItemChange)
		if err != nil {
			handleError(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
