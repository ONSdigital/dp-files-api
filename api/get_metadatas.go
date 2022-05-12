package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
)

type GetFilesMetadata func(ctx context.Context, collectionID string) ([]files.StoredRegisteredMetaData, error)

func HandlerGetFilesMetadata(getFilesMetadata GetFilesMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		collectionID := req.URL.Query().Get("collection_id")

		if collectionID == "" {
			err := errors.New("missing collection ID")
			writeError(w, buildErrors(err, "BadRequest"), http.StatusBadRequest)
			return
		}

		fm, err := getFilesMetadata(req.Context(), collectionID)
		if err != nil {
			handleError(w, err)
			return
		}

		fc := filesCollectionFromMetadata(fm)
		if err := respondWithFilesCollectionJSON(w, fc); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

type FilesCollection struct {
	Count      int64                            `json:"count"`
	Limit      int64                            `json:"limit"`
	Offset     int64                            `json:"offset"`
	TotalCount int64                            `json:"total_count"`
	Items      []files.StoredRegisteredMetaData `json:"items"`
}

func filesCollectionFromMetadata(f []files.StoredRegisteredMetaData) FilesCollection {
	count := int64(len(f))
	fc := FilesCollection{
		Count:      count,
		Limit:      count,
		Offset:     0,
		TotalCount: count,
		Items:      f,
	}
	return fc
}

func respondWithFilesCollectionJSON(w http.ResponseWriter, fc FilesCollection) error {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(fc)
	return err
}
