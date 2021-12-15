package steps

import (
	"context"
	"strconv"

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

func (c *FilesApiComponent) theFollowingDocumentShouldBeCreated(table *godog.Table) error {
	ctx := context.Background()

	metaData := files.StoredMetaData{}

	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaData{}, table)
	if err != nil {
		return err
	}

	expectedMetaData := keyValues.(*ExpectedMetaData)

	res := c.Mongo.Client.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
	assert.NoError(c.ApiFeature, res.Decode(&metaData))

	isPublishable, _ := strconv.ParseBool(expectedMetaData.IsPublishable)
	sizeInBytes, _ := strconv.ParseInt(expectedMetaData.SizeInBytes, 10, 64)
	assert.Equal(c.ApiFeature, isPublishable, metaData.IsPublishable)
	assert.Equal(c.ApiFeature, expectedMetaData.CollectionID, metaData.CollectionID)
	assert.Equal(c.ApiFeature, expectedMetaData.Title, metaData.Title)
	assert.Equal(c.ApiFeature, sizeInBytes, metaData.SizeInBytes)
	assert.Equal(c.ApiFeature, expectedMetaData.Type, metaData.Type)
	assert.Equal(c.ApiFeature, expectedMetaData.Licence, metaData.Licence)
	assert.Equal(c.ApiFeature, expectedMetaData.LicenceUrl, metaData.LicenceUrl)
	assert.Equal(c.ApiFeature, expectedMetaData.State, metaData.State)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasBeenRegistered(path string) error {
	ctx := context.Background()

	m := files.StoredMetaData{Path: path}
	_, err := c.Mongo.Client.Database("files").Collection("metadata").InsertOne(ctx, &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}
