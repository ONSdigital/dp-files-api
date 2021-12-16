package steps

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/messages-go/v16"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *FilesApiComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the file upload is registered with payload:`, c.iRegisterFile)
	ctx.Step(`^the following document entry should be created:`, c.theFollowingDocumentShouldBeCreated)
	ctx.Step(`^the file upload "([^"]*)" has been registered$`, c.theFileUploadHasBeenRegistered)
	ctx.Step(`^the file upload "([^"]*)" has been registered with:$`, c.theFileUploadHasBeenRegisteredWith)
	ctx.Step(`^the file upload "([^"]*)" is marked as complete with the etag "([^"]*)"$`, c.theFileUploadIsMarkedAsCompleteWithTheEtag)
	ctx.Step(`^the following document entry should be look like:$`, c.theFollowingDocumentEntryShouldBeLookLike)
	ctx.Step(`^the file upload "([^"]*)" has not been registered$`, c.theFileUploadHasNotBeenRegistered)


}

func (c *FilesApiComponent) iRegisterFile(payload *godog.DocString) error {
	return c.ApiFeature.IPostToWithBody("/v1/files", payload)
}

type ExpectedMetaData struct {
	Path          string
	IsPublishable string
	CollectionID  string
	Title         string
	SizeInBytes   string
	Type          string
	Licence       string
	LicenceUrl    string
	CreatedAt     string
	LastModified  string
	State         string
}

type ExpectedMetaDataUploadComplete struct {
	ExpectedMetaData
	Etag              string
	UploadCompletedAt string
}

func (c *FilesApiComponent) theFollowingDocumentShouldBeCreated(table *godog.Table) error {
	ctx := context.Background()

	metaData := files.StoredRegisteredMetaData{}

	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaData{}, table)
	if err != nil {
		return err
	}

	expectedMetaData := keyValues.(*ExpectedMetaData)

	res := c.Mongo.Client.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
	assert.NoError(c.ApiFeature, res.Decode(&metaData))

	isPublishable, _ := strconv.ParseBool(expectedMetaData.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(expectedMetaData.SizeInBytes, 10, 64)
	assert.Equal(c.ApiFeature, isPublishable, metaData.IsPublishable)
	assert.Equal(c.ApiFeature, expectedMetaData.CollectionID, metaData.CollectionID)
	assert.Equal(c.ApiFeature, expectedMetaData.Title, metaData.Title)
	assert.Equal(c.ApiFeature, sizeInBytes, metaData.SizeInBytes)
	assert.Equal(c.ApiFeature, expectedMetaData.Type, metaData.Type)
	assert.Equal(c.ApiFeature, expectedMetaData.Licence, metaData.Licence)
	assert.Equal(c.ApiFeature, expectedMetaData.LicenceUrl, metaData.LicenceUrl)
	assert.Equal(c.ApiFeature, expectedMetaData.State, metaData.State)
	assert.Equal(c.ApiFeature, expectedMetaData.CreatedAt, metaData.CreatedAt.Format(time.RFC3339))
	assert.Equal(c.ApiFeature, expectedMetaData.LastModified, metaData.LastModified.Format(time.RFC3339))

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasBeenRegistered(path string) error {
	ctx := context.Background()

	m := files.StoredRegisteredMetaData{Path: path}

	_, err := c.Mongo.Client.Database("files").Collection("metadata").InsertOne(ctx, &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasBeenRegisteredWith(path string, table *godog.Table) error {
	ctx := context.Background()
	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaData{}, table)
	if err != nil {
		return err
	}

	data := keyValues.(*ExpectedMetaData)

	isPublishable, _ := strconv.ParseBool(data.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(data.SizeInBytes, 10, 64)
	createdAt, _ := time.Parse(time.RFC3339, data.CreatedAt)
	lastModified, _ := time.Parse(time.RFC3339, data.LastModified)

	m := files.StoredRegisteredMetaData{
		Path:          path,
		IsPublishable: isPublishable,
		CollectionID:  data.CollectionID,
		Title:         data.Title,
		SizeInBytes:   sizeInBytes,
		Type:          data.Type,
		Licence:       data.Licence,
		LicenceUrl:    data.LicenceUrl,
		State:         data.State,
		CreatedAt:     createdAt,
		LastModified:  lastModified,
	}
	_, err = c.Mongo.Client.Database("files").Collection("metadata").InsertOne(ctx, &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasNotBeenRegistered(path string) error {
	ctx := context.Background()
	_, err := c.Mongo.Client.Database("files").Collection("metadata").DeleteMany(ctx, bson.M{"path": path})

	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}


func (c *FilesApiComponent) theFileUploadIsMarkedAsCompleteWithTheEtag(path, etag string) error {
	json := fmt.Sprintf(`{
	"path": "%s",
	"etag": "%s"
}`, path, etag)
	payload := messages.PickleDocString{
		MediaType: "application/json",
		Content:   json,
	}
	return c.ApiFeature.IPostToWithBody("/v1/files/upload-complete", &payload)
}

func (c *FilesApiComponent) theFollowingDocumentEntryShouldBeLookLike(table *godog.Table) error {
	ctx := context.Background()

	metaData := files.StoredRegisteredMetaData{}

	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaDataUploadComplete{}, table)
	if err != nil {
		return err
	}

	expectedMetaData := keyValues.(*ExpectedMetaDataUploadComplete)

	res := c.Mongo.Client.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
	assert.NoError(c.ApiFeature, res.Decode(&metaData))

	isPublishable, _ := strconv.ParseBool(expectedMetaData.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(expectedMetaData.SizeInBytes, 10, 64)
	assert.Equal(c.ApiFeature, isPublishable, metaData.IsPublishable)
	assert.Equal(c.ApiFeature, expectedMetaData.CollectionID, metaData.CollectionID)
	assert.Equal(c.ApiFeature, expectedMetaData.Title, metaData.Title)
	assert.Equal(c.ApiFeature, sizeInBytes, metaData.SizeInBytes)
	assert.Equal(c.ApiFeature, expectedMetaData.Type, metaData.Type)
	assert.Equal(c.ApiFeature, expectedMetaData.Licence, metaData.Licence)
	assert.Equal(c.ApiFeature, expectedMetaData.LicenceUrl, metaData.LicenceUrl)
	assert.Equal(c.ApiFeature, expectedMetaData.State, metaData.State)
	assert.Equal(c.ApiFeature, expectedMetaData.Etag, metaData.Etag)
	assert.Equal(c.ApiFeature, expectedMetaData.CreatedAt, metaData.CreatedAt.Format(time.RFC3339))
	assert.Equal(c.ApiFeature, expectedMetaData.LastModified, metaData.LastModified.Format(time.RFC3339))
	assert.Equal(c.ApiFeature, expectedMetaData.UploadCompletedAt, metaData.UploadCompletedAt.Format(time.RFC3339))

	return c.ApiFeature.StepError()
}
