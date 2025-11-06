package steps

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type FileEventData struct {
	RequestedByID string
	Action        string
	Resource      string
	FilePath      string
	BundleID      string
	CollectionID  string
}

func (c *FilesAPIComponent) iCreateFileEvent(payload *godog.DocString) error {
	return c.APIFeature.IPostToWithBody("/file-events", payload)
}

func (c *FilesAPIComponent) theFileEventShouldBeCreatedInTheDatabase() error {
	ctx := context.Background()

	count, err := c.mongoClient.Database("files").Collection("file_events").CountDocuments(ctx, bson.M{})
	assert.NoError(c.APIFeature, err)
	assert.Greater(c.APIFeature, count, int64(0), "Expected at least one file event in the database")

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFollowingFileEventsExistInTheDatabase(table *godog.Table) error {
	ctx := context.Background()
	collection := c.mongoClient.Database("files").Collection("file_events")

	baseTime := time.Date(2025, 10, 28, 12, 0, 0, 0, time.UTC)

	assist := assistdog.NewDefault()
	events, err := assist.CreateSlice(&FileEventData{}, table)
	if err != nil {
		return err
	}

	for i, item := range events.([]*FileEventData) {
		fileObj := bson.M{
			"path":           item.FilePath,
			"is_publishable": true,
			"title":          "Test File",
			"size_in_bytes":  1024,
			"type":           "text/csv",
			"licence":        "OGL v3",
			"licence_url":    "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
		}

		if item.BundleID != "" {
			fileObj["bundle_id"] = item.BundleID
		}

		createdAt := baseTime.Add(-time.Duration(i+1) * time.Hour)

		event := bson.M{
			"created_at": createdAt,
			"requested_by": bson.M{
				"id":    item.RequestedByID,
				"email": item.RequestedByID + "@example.com",
			},
			"action":   item.Action,
			"resource": item.Resource,
			"file":     fileObj,
		}

		_, err := collection.InsertOne(ctx, event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *FilesAPIComponent) theResponseShouldContainFileEvents(expectedCount string) error {
	var eventsList files.EventsList

	body, err := io.ReadAll(c.APIFeature.HTTPResponse.Body)
	assert.NoError(c.APIFeature, err)

	err = json.Unmarshal(body, &eventsList)
	assert.NoError(c.APIFeature, err)

	var expected int
	switch expectedCount {
	case "5":
		expected = 5
	case "2":
		expected = 2
	case "1":
		expected = 1
	case "0":
		expected = 0
	default:
		expected = 0
	}

	assert.Equal(c.APIFeature, expected, eventsList.Count, "Expected %d file events but got %d", expected, eventsList.Count)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theResponseShouldContainAtLeastFileEvent(minCount string) error {
	var eventsList files.EventsList

	body, err := io.ReadAll(c.APIFeature.HTTPResponse.Body)
	assert.NoError(c.APIFeature, err)

	err = json.Unmarshal(body, &eventsList)
	assert.NoError(c.APIFeature, err)

	var expectedMin int
	switch minCount {
	case "1":
		expectedMin = 1
	default:
		expectedMin = 1
	}

	assert.GreaterOrEqual(c.APIFeature, eventsList.Count, expectedMin, "Expected at least %d file event(s)", expectedMin)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theResponseShouldHavePaginationWithLimitAndOffset(limit, offset string) error {
	var eventsList files.EventsList

	body, err := io.ReadAll(c.APIFeature.HTTPResponse.Body)
	assert.NoError(c.APIFeature, err)

	err = json.Unmarshal(body, &eventsList)
	assert.NoError(c.APIFeature, err)

	var expectedLimit, expectedOffset int
	switch limit {
	case "20":
		expectedLimit = 20
	case "2":
		expectedLimit = 2
	default:
		expectedLimit = 20
	}

	switch offset {
	case "0":
		expectedOffset = 0
	case "1":
		expectedOffset = 1
	default:
		expectedOffset = 0
	}

	assert.Equal(c.APIFeature, expectedLimit, eventsList.Limit, "Expected limit %d but got %d", expectedLimit, eventsList.Limit)
	assert.Equal(c.APIFeature, expectedOffset, eventsList.Offset, "Expected offset %d but got %d", expectedOffset, eventsList.Offset)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) allReturnedEventsShouldHaveFilePath(expectedPath string) error {
	var eventsList files.EventsList

	body, err := io.ReadAll(c.APIFeature.HTTPResponse.Body)
	assert.NoError(c.APIFeature, err)

	err = json.Unmarshal(body, &eventsList)
	assert.NoError(c.APIFeature, err)

	for _, event := range eventsList.Items {
		assert.Equal(c.APIFeature, expectedPath, event.File.Path, "Expected all events to have path %s", expectedPath)
	}

	return c.APIFeature.StepError()
}
