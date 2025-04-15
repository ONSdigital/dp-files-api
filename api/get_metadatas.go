package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/log.go/v2/log"
)

type GetFilesMetadata func(ctx context.Context, collectionID, bundleID string) ([]files.StoredRegisteredMetaData, error)

func HandlerGetFilesMetadata(getFilesMetadata GetFilesMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		collectionID := req.URL.Query().Get("collection_id")
		bundleID := req.URL.Query().Get("bundle_id")

		if collectionID == "" && bundleID == "" {
			err := errors.New("missing required ID: either collection_id or bundle_id must be provided")
			writeError(w, buildErrors(err, "BadRequest"), http.StatusBadRequest)
			return
		}

		if collectionID != "" && bundleID != "" {
			err := errors.New("only one of collection_id or bundle_id should be provided")
			writeError(w, buildErrors(err, "BadRequest"), http.StatusBadRequest)
			return
		}

		fm, err := getFilesMetadata(req.Context(), collectionID, bundleID)
		if err != nil {
			idType := "collection"
			idValue := collectionID
			if bundleID != "" {
				idType = "bundle"
				idValue = bundleID
			}
			log.Error(req.Context(), "file metadata fetch failed", err, log.Data{idType: idValue})
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
