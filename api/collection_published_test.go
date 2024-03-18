package api_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/stretchr/testify/assert"
)

func TestPublishHandlerHandlesUnexpectedPublishingError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/ignore.txt", strings.NewReader(`{"collection_id": "asdfghjkl"}`))

	h := api.HandleMarkCollectionPublished(func(ctx context.Context, collectionID string) error {
		return errors.New("broken")
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}

func TestPublishHandlerDontReturnErrorIfCollectionEmpty(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/files/ignore.txt", strings.NewReader(`{"collection_id": "asdfghjkl"}`))

	h := api.HandleMarkCollectionPublished(func(ctx context.Context, collectionID string) error {
		return store.ErrNoFilesInCollection
	})

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}
