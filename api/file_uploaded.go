package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/go-playground/validator"
)

type MarkUploadComplete func(ctx context.Context, metaData files.FileEtagChange) error

func HandleMarkUploadComplete(markUploaded MarkUploadComplete) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ec, err := getEtagChangeFromRequest(req)
		if err != nil {
			writeError(w, buildErrors(err, "BadJsonEncoding"), http.StatusBadRequest)
			return
		}

		if err = validator.New().Struct(ec); err != nil {
			handleError(w, err)
			return
		}

		if err := markUploaded(req.Context(), generateFileEtagChange(ec, mux.Vars(req)["path"])); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getEtagChangeFromRequest(req *http.Request) (EtagChange, error) {
	ec := EtagChange{}
	err := json.NewDecoder(req.Body).Decode(&ec)
	return ec, err
}
