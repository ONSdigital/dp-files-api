package store

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-files-api/sdk"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
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

// GetFileEvents retrieves file events with optional filters and pagination
func (store *Store) GetFileEvents(ctx context.Context, limit, offset int, path string, after, before *time.Time) (*sdk.EventsList, error) {
	filter := bson.M{}

	if path != "" {
		filter["file.path"] = path
	}

	if after != nil || before != nil {
		createdAtFilter := bson.M{}
		if after != nil {
			createdAtFilter["$gte"] = after
		}
		if before != nil {
			createdAtFilter["$lte"] = before
		}
		filter["created_at"] = createdAtFilter
	}

	if path != "" {
		count, err := store.fileEventsCollection.Count(ctx, bson.M{"file.path": path})
		if err != nil {
			log.Error(ctx, "failed to check if path exists", err, log.Data{"path": path})
			return nil, err
		}
		if count == 0 {
			return nil, ErrPathNotFound
		}
	}

	totalCount, err := store.fileEventsCollection.Count(ctx, filter)
	if err != nil {
		log.Error(ctx, "failed to count file events", err)
		return nil, err
	}

	events := make([]sdk.FileEvent, 0)

	_, err = store.fileEventsCollection.Find(ctx, filter, &events,
		mongodriver.Sort(bson.M{"created_at": -1}),
		mongodriver.Offset(offset),
		mongodriver.Limit(limit),
	)
	if err != nil {
		log.Error(ctx, "failed to find file events", err)
		return nil, err
	}

	eventsList := &sdk.EventsList{
		Count:      len(events),
		Limit:      limit,
		Offset:     offset,
		TotalCount: totalCount,
		Items:      events,
	}

	return eventsList, nil
}
