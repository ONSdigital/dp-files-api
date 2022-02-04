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
	ctx.Step(`^the file upload "([^"]*)" has been completed with:$`, c.theFileUploadHasBeenCompletedWith)
	ctx.Step(`^the file upload "([^"]*)" is marked as complete with the etag "([^"]*)"$`, c.theFileUploadIsMarkedAsCompleteWithTheEtag)
	ctx.Step(`^the following document entry should be look like:$`, c.theFollowingDocumentEntryShouldBeLookLike)
	ctx.Step(`^the file upload "([^"]*)" has not been registered$`, c.theFileUploadHasNotBeenRegistered)
	ctx.Step(`^the file metadata is requested for the file "([^"]*)"$`, c.theFileMetadataIsRequested)
	ctx.Step(`^the file "([^"]*)" has not been registered$`, c.theFileHasNotBeenRegistered)
	ctx.Step(`^I publish the collection "([^"]*)"$`, c.iPublishTheCollection)
	ctx.Step(`^the file "([^"]*)" is marked as decrypted with etag "([^"]*)"$`, c.theFileIsMarkedAsDecrypted)
	ctx.Step(`^the file upload "([^"]*)" has been published with:$`, c.theFileUploadHasBeenPublishedWith)
}

func (c *FilesApiComponent) iRegisterFile(payload *godog.DocString) error {
	return c.ApiFeature.IPostToWithBody("/v1/files/register", payload)
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

type ExpectedMetaDataPublished struct {
	ExpectedMetaDataUploadComplete
	PublishedAt string
}

type ExpectedMetaDataDecrypted struct {
	ExpectedMetaDataPublished
	DecryptedAt string
}

func (c *FilesApiComponent) theFileHasNotBeenRegistered(arg1 string) error {
	return nil
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

	res := c.mongoClient.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
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

	_, err := c.mongoClient.Database("files").Collection("metadata").InsertOne(ctx, &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasBeenPublishedWith(path string, table *godog.Table) error {
	keyValues, err := assistdog.NewDefault().CreateInstance(&ExpectedMetaDataPublished{}, table)
	if err != nil {
		return err
	}

	data := keyValues.(*ExpectedMetaDataPublished)

	isPublishable, _ := strconv.ParseBool(data.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(data.SizeInBytes, 10, 64)
	createdAt, _ := time.Parse(time.RFC3339, data.CreatedAt)
	lastModified, _ := time.Parse(time.RFC3339, data.LastModified)
	uploadCompletedAt, _ := time.Parse(time.RFC3339, data.UploadCompletedAt)
	publishedAt, _ := time.Parse(time.RFC3339, data.PublishedAt)

	m := files.StoredRegisteredMetaData{
		Path:              path,
		IsPublishable:     isPublishable,
		CollectionID:      data.CollectionID,
		Title:             data.Title,
		SizeInBytes:       sizeInBytes,
		Type:              data.Type,
		Licence:           data.Licence,
		LicenceUrl:        data.LicenceUrl,
		State:             data.State,
		CreatedAt:         createdAt,
		LastModified:      lastModified,
		UploadCompletedAt: uploadCompletedAt,
		PublishedAt:       publishedAt,
		Etag:              data.Etag,
	}

	_, err = c.mongoClient.Database("files").Collection("metadata").InsertOne(context.Background(), &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasBeenCompletedWith(path string, table *godog.Table) error {
	keyValues, err := assistdog.NewDefault().CreateInstance(&ExpectedMetaDataUploadComplete{}, table)
	if err != nil {
		return err
	}

	data := keyValues.(*ExpectedMetaDataUploadComplete)

	isPublishable, _ := strconv.ParseBool(data.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(data.SizeInBytes, 10, 64)
	createdAt, _ := time.Parse(time.RFC3339, data.CreatedAt)
	lastModified, _ := time.Parse(time.RFC3339, data.LastModified)
	uploadCompletedAt, _ := time.Parse(time.RFC3339, data.UploadCompletedAt)

	m := files.StoredRegisteredMetaData{
		Path:              path,
		IsPublishable:     isPublishable,
		CollectionID:      data.CollectionID,
		Title:             data.Title,
		SizeInBytes:       sizeInBytes,
		Type:              data.Type,
		Licence:           data.Licence,
		LicenceUrl:        data.LicenceUrl,
		State:             data.State,
		CreatedAt:         createdAt,
		LastModified:      lastModified,
		UploadCompletedAt: uploadCompletedAt,
		Etag:              data.Etag,
	}

	_, err = c.mongoClient.Database("files").Collection("metadata").InsertOne(context.Background(), &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasBeenRegisteredWith(path string, table *godog.Table) error {
	keyValues, err := assistdog.NewDefault().CreateInstance(&ExpectedMetaData{}, table)
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

	_, err = c.mongoClient.Database("files").Collection("metadata").InsertOne(context.Background(), &m)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileUploadHasNotBeenRegistered(path string) error {
	ctx := context.Background()
	_, err := c.mongoClient.Database("files").Collection("metadata").DeleteMany(ctx, bson.M{"path": path})

	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) theFileIsMarkedAsDecrypted(path, etag string) error {
	json := fmt.Sprintf(`{"path": "%s","etag": "%s"}`, path, etag)
	return c.ApiFeature.IPostToWithBody("/v1/files/decrypted", &messages.PickleDocString{Content: json})
}

func (c *FilesApiComponent) theFileUploadIsMarkedAsCompleteWithTheEtag(path, etag string) error {
	json := fmt.Sprintf(`{
	"path": "%s",
	"etag": "%s"
}`, path, etag)
	return c.ApiFeature.IPostToWithBody("/v1/files/upload-complete", &messages.PickleDocString{Content: json})
}

func (c *FilesApiComponent) theFileMetadataIsRequested(filepath string) error {
	return c.ApiFeature.IGet(fmt.Sprintf("/v1/files/%s", filepath))
}

func (c *FilesApiComponent) theFollowingDocumentEntryShouldBeLookLike(table *godog.Table) error {
	ctx := context.Background()

	metaData := files.StoredRegisteredMetaData{}

	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaDataDecrypted{}, table)
	if err != nil {
		return err
	}

	expectedMetaData := keyValues.(*ExpectedMetaDataDecrypted)

	res := c.mongoClient.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
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
	assert.Equal(c.ApiFeature, expectedMetaData.CreatedAt, metaData.CreatedAt.Format(time.RFC3339), "CREATED AT")
	assert.Equal(c.ApiFeature, expectedMetaData.LastModified, metaData.LastModified.Format(time.RFC3339), "LAST MODIFIED")
	if expectedMetaData.PublishedAt != "" {
		assert.Equal(c.ApiFeature, expectedMetaData.PublishedAt, metaData.PublishedAt.Format(time.RFC3339), "DECRYPTED AT")
	}
	if expectedMetaData.DecryptedAt != "" {
		assert.Equal(c.ApiFeature, expectedMetaData.DecryptedAt, metaData.DecryptedAt.Format(time.RFC3339), "DECRYPTED AT")
	}

	return c.ApiFeature.StepError()
}

func (c *FilesApiComponent) iPublishTheCollection(collectionID string) error {
	body := fmt.Sprintf(`{"collection_id": "%s"}`, collectionID)
	c.ApiFeature.IPostToWithBody("/v1/files/publish", &messages.PickleDocString{MediaType: "application/json", Content: body})

	return c.ApiFeature.StepError()
}
