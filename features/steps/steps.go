package steps

import (
	"context"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"strconv"
	"strings"
)

func (c *FilesApiComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the file upload is registered with payload:`, c.iRegisterFile)
	ctx.Step(`^the following document entry should be created:`, c.theFollowingDocumentShouldBeCreated)
}

func (c *FilesApiComponent) iShouldReceiveAHelloworldResponse() error {
	responseBody := c.ApiFeature.HttpResponse.Body
	body, _ := ioutil.ReadAll(responseBody)

	assert.Equal(c.ApiFeature, `{"message":"Hello, World!"}`, strings.TrimSpace(string(body)))

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) iRegisterFile(payload *godog.DocString) error {

	err := c.ApiFeature.IPostToWithBody("/v1/files", payload)
	if err != nil {
		return err
	}

	return nil
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

	metaData := files.MetaData{}

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
