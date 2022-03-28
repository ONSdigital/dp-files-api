package api

import (
	"github.com/ONSdigital/dp-files-api/files"
)

type EtagChange struct {
	Etag string `json:"etag" validate:"required"`
}

func generateFileEtagChange(m EtagChange, path string) files.FileEtagChange {
	return files.FileEtagChange{
		Path: path,
		Etag: m.Etag,
	}
}
