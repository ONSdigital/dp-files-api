package steps

import (
	"context"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

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
