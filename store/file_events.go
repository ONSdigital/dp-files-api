package store

import (
	"context"

	"github.com/ONSdigital/dp-files-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// CreateFileEvent inserts a new file event into the file_events collection
func (store *Store) CreateFileEvent(ctx context.Context, event *sdk.FileEvent) error {
	now := store.clock.GetCurrentTime()
	event.CreatedAt = &now

	_, err := store.fileEventsCollection.Insert(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to insert file event", err, log.Data{
			"action":   event.Action,
			"resource": event.Resource,
		})
		return err
	}

	return nil
}
