package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-files-api/store"

	"github.com/ONSdigital/dp-files-api/config"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"

	messages "github.com/cucumber/messages/go/v21"

	"github.com/ONSdigital/dp-files-api/files"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *FilesAPIComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I am an authorised user$`, c.iAmAnAuthorisedUser)
	ctx.Step(`^I am not an authorised user$`, c.iAmNotAnAuthorisedUser)
	ctx.Step(`^the file upload is registered with payload:`, c.iRegisterFile)
	ctx.Step(`^the following document entry should be created:`, c.theFollowingDocumentShouldBeCreated)
	ctx.Step(`^the file upload "([^"]*)" has been registered$`, c.theFileUploadHasBeenRegistered)
	ctx.Step(`^the file upload "([^"]*)" has been registered with:$`, c.theFileUploadHasBeenRegisteredWith)
	ctx.Step(`^the file upload "([^"]*)" has been completed with:$`, c.theFileUploadHasBeenCompletedWith)
	ctx.Step(`^the file upload "([^"]*)" is marked as complete with the etag "([^"]*)"$`, c.theFileUploadIsMarkedAsCompleteWithTheEtag)
	ctx.Step(`^the file "([^"]*)" is marked as published$`, c.theFileIsMarkedAsPublished)
	ctx.Step(`^the file "([^"]*)" is marked as moved with etag "([^"]*)"$`, c.theFileIsMarkedAsMoved)
	ctx.Step(`^the following document entry should look like:$`, c.theFollowingDocumentEntryShouldLookLike)
	ctx.Step(`^the file upload "([^"]*)" has not been registered$`, c.theFileUploadHasNotBeenRegistered)
	ctx.Step(`^the file metadata is requested for the file "([^"]*)"$`, c.theFileMetadataIsRequested)
	ctx.Step(`^the file "([^"]*)" has not been registered$`, c.theFileHasNotBeenRegistered)
	ctx.Step(`^I publish the collection "([^"]*)"$`, c.iPublishTheCollection)
	ctx.Step(`^I publish the bundle "([^"]*)"$`, c.iPublishTheBundle)
	ctx.Step(`^the file upload "([^"]*)" has been published with:$`, c.theFileUploadHasBeenPublishedWith)
	ctx.Step(`^the following PUBLISHED message is sent to Kakfa:$`, c.theFollowingPublishedMessageIsSent)
	ctx.Step(`^Kafka Consumer Group is running$`, c.kafkaConsumerGroupIsRunning)
	ctx.Step(`^I am in web mode$`, c.iAmInWebMode)
	ctx.Step(`^I set the collection ID to "([^"]*)" for file "([^"]*)"$`, c.iSetTheCollectionIDToForFile)
	ctx.Step(`^I get files in the collection "([^"]*)"$`, c.iGetFilesInTheCollection)
	ctx.Step(`^I set the bundle ID to "([^"]*)" for file "([^"]*)"$`, c.iSetTheBundleIDToForFile)
	ctx.Step(`^I get files in the bundle "([^"]*)"$`, c.iGetFilesInTheBundle)
	ctx.Step(`^I get files with both collection_id "([^"]*)" and bundle_id "([^"]*)"$`, c.iGetFilesWithBothCollectionAndBundleID)
	ctx.Step(`^the file upload "([^"]*)" is removed$`, c.theFileUploadIsRemoved)
	ctx.Step(`^I create a file event with payload:$`, c.iCreateFileEvent)
	ctx.Step(`^the file event should be created in the database$`, c.theFileEventShouldBeCreatedInTheDatabase)
	ctx.Step(`^the following file events exist in the database:$`, c.theFollowingFileEventsExistInTheDatabase)
	ctx.Step(`^the response should contain "([^"]*)" file events$`, c.theResponseShouldContainFileEvents)
	ctx.Step(`^the response should contain at least "([^"]*)" file event$`, c.theResponseShouldContainAtLeastFileEvent)
	ctx.Step(`^the response should have pagination with limit "([^"]*)" and offset "([^"]*)"$`, c.theResponseShouldHavePaginationWithLimitAndOffset)
	ctx.Step(`^all returned events should have file path "([^"]*)"$`, c.allReturnedEventsShouldHaveFilePath)
	ctx.Step(`^I update the content item of the file "([^"]*)" with:`, c.iUpdateTheContentItemOfTheFileWith)
}

func (c *FilesAPIComponent) iAmAnAuthorisedUser() error {
	c.isAuthorised = true
	if c.APIFeature != nil {
		return c.APIFeature.ISetTheHeaderTo("Authorization", "Bearer test-token")
	}
	return nil
}

func (c *FilesAPIComponent) iAmNotAnAuthorisedUser() error {
	c.isAuthorised = false
	if c.APIFeature != nil {
		return c.APIFeature.ISetTheHeaderTo("Authorization", "Bearer test-token")
	}
	return nil
}

func (c *FilesAPIComponent) iRegisterFile(payload *godog.DocString) error {
	return c.APIFeature.IPostToWithBody("/files", payload)
}

type ExpectedMetaData struct {
	Path          string
	IsPublishable string
	CollectionID  string
	BundleID      string
	Title         string
	SizeInBytes   string
	Type          string
	Licence       string
	LicenceURL    string
	CreatedAt     string
	LastModified  string
	State         string
	DatasetID     string
	Edition       string
	Version       string
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

type ExpectedMetaDataMoved struct {
	ExpectedMetaDataPublished
	MovedAt string
}

func (c *FilesAPIComponent) theFileHasNotBeenRegistered(arg1 string) error {
	return nil
}

func (c *FilesAPIComponent) theFollowingDocumentShouldBeCreated(table *godog.Table) error {
	ctx := context.Background()

	metaData := files.StoredRegisteredMetaData{}

	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaData{}, table)
	if err != nil {
		return err
	}

	expectedMetaData := keyValues.(*ExpectedMetaData)

	res := c.mongoClient.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
	assert.NoError(c.APIFeature, res.Decode(&metaData))

	isPublishable, _ := strconv.ParseBool(expectedMetaData.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(expectedMetaData.SizeInBytes, 10, 64)
	assert.Equal(c.APIFeature, isPublishable, metaData.IsPublishable)
	if expectedMetaData.CollectionID == "" {
		assert.Nil(c.APIFeature, metaData.CollectionID)
	} else {
		assert.Equal(c.APIFeature, expectedMetaData.CollectionID, *metaData.CollectionID)
	}
	if expectedMetaData.BundleID == "" {
		assert.Nil(c.APIFeature, metaData.BundleID)
	} else {
		assert.Equal(c.APIFeature, expectedMetaData.BundleID, *metaData.BundleID)
	}

	if expectedMetaData.DatasetID == "" && expectedMetaData.Edition == "" && expectedMetaData.Version == "" {
		assert.Nil(c.APIFeature, metaData.ContentItem)
	} else {
		assert.NotNil(c.APIFeature, metaData.ContentItem)
		assert.Equal(c.APIFeature, expectedMetaData.DatasetID, metaData.ContentItem.DatasetID)
		assert.Equal(c.APIFeature, expectedMetaData.Edition, metaData.ContentItem.Edition)
		assert.Equal(c.APIFeature, expectedMetaData.Version, metaData.ContentItem.Version)
	}

	assert.Equal(c.APIFeature, expectedMetaData.Title, metaData.Title)
	assert.Equal(c.APIFeature, sizeInBytes, metaData.SizeInBytes)
	assert.Equal(c.APIFeature, expectedMetaData.Type, metaData.Type)
	assert.Equal(c.APIFeature, expectedMetaData.Licence, metaData.Licence)
	assert.Equal(c.APIFeature, expectedMetaData.LicenceURL, metaData.LicenceURL)
	assert.Equal(c.APIFeature, expectedMetaData.State, metaData.State)
	assert.Equal(c.APIFeature, expectedMetaData.CreatedAt, metaData.CreatedAt.Format(time.RFC3339))
	assert.Equal(c.APIFeature, expectedMetaData.LastModified, metaData.LastModified.Format(time.RFC3339))

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFileUploadHasBeenRegistered(path string) error {
	ctx := context.Background()

	m := files.StoredRegisteredMetaData{Path: path}

	_, err := c.mongoClient.Database("files").Collection("metadata").InsertOne(ctx, &m)
	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFileUploadHasBeenPublishedWith(path string, table *godog.Table) error {
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
		CollectionID:      &data.CollectionID,
		BundleID:          &data.BundleID,
		Title:             data.Title,
		SizeInBytes:       sizeInBytes,
		Type:              data.Type,
		Licence:           data.Licence,
		LicenceURL:        data.LicenceURL,
		State:             data.State,
		CreatedAt:         createdAt,
		LastModified:      lastModified,
		UploadCompletedAt: &uploadCompletedAt,
		PublishedAt:       &publishedAt,
		Etag:              data.Etag,
	}

	_, err = c.mongoClient.Database("files").Collection("metadata").InsertOne(context.Background(), &m)
	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFileUploadHasBeenCompletedWith(path string, table *godog.Table) error {
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
		Title:             data.Title,
		SizeInBytes:       sizeInBytes,
		Type:              data.Type,
		Licence:           data.Licence,
		LicenceURL:        data.LicenceURL,
		State:             data.State,
		CreatedAt:         createdAt,
		LastModified:      lastModified,
		UploadCompletedAt: &uploadCompletedAt,
		Etag:              data.Etag,
	}

	if data.CollectionID != "" {
		m.CollectionID = &data.CollectionID
	}

	if data.BundleID != "" {
		m.BundleID = &data.BundleID
	}

	_, err = c.mongoClient.Database("files").Collection("metadata").InsertOne(context.Background(), &m)
	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFileUploadHasBeenRegisteredWith(path string, table *godog.Table) error {
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
		Title:         data.Title,
		SizeInBytes:   sizeInBytes,
		Type:          data.Type,
		Licence:       data.Licence,
		LicenceURL:    data.LicenceURL,
		State:         data.State,
		CreatedAt:     createdAt,
		LastModified:  lastModified,
	}

	if data.CollectionID != "" {
		m.CollectionID = &data.CollectionID
	}

	if data.BundleID != "" {
		m.BundleID = &data.BundleID
	}

	if data.DatasetID != "" || data.Edition != "" || data.Version != "" {
		m.ContentItem = &files.StoredContentItem{
			DatasetID: data.DatasetID,
			Edition:   data.Edition,
			Version:   data.Version,
		}
	}

	_, err = c.mongoClient.Database("files").Collection("metadata").InsertOne(context.Background(), &m)
	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFileUploadHasNotBeenRegistered(path string) error {
	ctx := context.Background()
	_, err := c.mongoClient.Database("files").Collection("metadata").DeleteMany(ctx, bson.M{"path": path})

	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFileIsMarkedAsMoved(path, etag string) error {
	json := fmt.Sprintf(`{"etag": %q, "state": %q}`, etag, store.StateMoved)
	return c.APIFeature.IPatch(fmt.Sprintf("/files/%s", path), &messages.PickleDocString{Content: json})
}

func (c *FilesAPIComponent) theFileIsMarkedAsPublished(path string) error {
	json := fmt.Sprintf(`{"state": %q}`, store.StatePublished)
	return c.APIFeature.IPatch(fmt.Sprintf("/files/%s", path), &messages.PickleDocString{Content: json})
}

func (c *FilesAPIComponent) theFileUploadIsMarkedAsCompleteWithTheEtag(path, etag string) error {
	json := fmt.Sprintf(`{
	"etag": "%s",
	"state": "%s"
}`, etag, store.StateUploaded)
	return c.APIFeature.IPatch(fmt.Sprintf("/files/%s", path), &messages.PickleDocString{Content: json})
}

func (c *FilesAPIComponent) iSetTheCollectionIDToForFile(collectionID, path string) error {
	json := fmt.Sprintf(`{"collection_id": %q}`, collectionID)
	return c.APIFeature.IPatch(fmt.Sprintf("/files/%s", path), &messages.PickleDocString{Content: json})
}

func (c *FilesAPIComponent) theFileMetadataIsRequested(filepath string) error {
	return c.APIFeature.IGet(fmt.Sprintf("/files/%s", filepath))
}

func (c *FilesAPIComponent) theFileUploadIsRemoved(path string) error {
	return c.APIFeature.IDelete(fmt.Sprintf("/files/%s", path))
}

func (c *FilesAPIComponent) theFollowingDocumentEntryShouldLookLike(table *godog.Table) error {
	ctx := context.Background()

	metaData := files.StoredRegisteredMetaData{}

	assist := assistdog.NewDefault()
	keyValues, err := assist.CreateInstance(&ExpectedMetaDataMoved{}, table)
	if err != nil {
		return err
	}

	expectedMetaData := keyValues.(*ExpectedMetaDataMoved)

	_ = c.APIFeature.IGet(fmt.Sprintf("/files/%s", expectedMetaData.Path))
	responseBody := c.APIFeature.HTTPResponse.Body
	body, _ := io.ReadAll(responseBody)
	assert.NoError(c.APIFeature, json.Unmarshal(body, &metaData))

	dbMetadata := files.StoredRegisteredMetaData{}
	res := c.mongoClient.Database("files").Collection("metadata").FindOne(ctx, bson.M{"path": expectedMetaData.Path})
	assert.NoError(c.APIFeature, res.Decode(&dbMetadata))

	metaData.CreatedAt = dbMetadata.CreatedAt
	metaData.LastModified = dbMetadata.LastModified
	metaData.PublishedAt = dbMetadata.PublishedAt
	metaData.MovedAt = dbMetadata.MovedAt

	if metaData.CollectionID != nil && metaData.State == store.StatePublished {
		dbCollection := files.StoredCollection{}
		res = c.mongoClient.Database("files").Collection("collections").FindOne(ctx, bson.M{"id": *metaData.CollectionID})
		_ = res.Decode(&dbCollection)

		if dbCollection.State == store.StatePublished {
			metaData.LastModified = dbCollection.LastModified
			metaData.PublishedAt = dbCollection.PublishedAt
		}
	}

	if metaData.BundleID != nil && metaData.State == store.StatePublished {
		dbBundle := files.StoredBundle{}
		res = c.mongoClient.Database("files").Collection("bundles").FindOne(ctx, bson.M{"id": *metaData.BundleID})
		_ = res.Decode(&dbBundle)

		if dbBundle.State == store.StatePublished {
			metaData.LastModified = dbBundle.LastModified
			// metaData.PublishedAt = dbBundle.PublishedAt // TODO: uncomment when PublishedAt is added to StoredBundle struct
		}
	}

	isPublishable, _ := strconv.ParseBool(expectedMetaData.IsPublishable)
	sizeInBytes, _ := strconv.ParseUint(expectedMetaData.SizeInBytes, 10, 64)
	assert.Equal(c.APIFeature, isPublishable, metaData.IsPublishable)
	if expectedMetaData.CollectionID != "" {
		assert.Equal(c.APIFeature, expectedMetaData.CollectionID, *metaData.CollectionID)
	}
	if expectedMetaData.BundleID != "" {
		assert.Equal(c.APIFeature, expectedMetaData.BundleID, *metaData.BundleID)
	}
	assert.Equal(c.APIFeature, expectedMetaData.Title, metaData.Title)
	assert.Equal(c.APIFeature, sizeInBytes, metaData.SizeInBytes)
	assert.Equal(c.APIFeature, expectedMetaData.Type, metaData.Type)
	assert.Equal(c.APIFeature, expectedMetaData.Licence, metaData.Licence)
	assert.Equal(c.APIFeature, expectedMetaData.LicenceURL, metaData.LicenceURL)
	assert.Equal(c.APIFeature, expectedMetaData.State, metaData.State)
	assert.Equal(c.APIFeature, expectedMetaData.Etag, metaData.Etag)
	assert.Equal(c.APIFeature, expectedMetaData.CreatedAt, metaData.CreatedAt.Format(time.RFC3339), "CREATED AT")
	assert.Equal(c.APIFeature, expectedMetaData.LastModified, metaData.LastModified.Format(time.RFC3339), "LAST MODIFIED")
	// TODO: remove expectedMetaData.BundleID == "" check once PublishedAt is added to StoredBundle struct
	if expectedMetaData.PublishedAt != "" && expectedMetaData.BundleID == "" {
		assert.Equal(c.APIFeature, expectedMetaData.PublishedAt, metaData.PublishedAt.Format(time.RFC3339), "PUBLISHED AT")
	}
	if expectedMetaData.MovedAt != "" {
		assert.Equal(c.APIFeature, expectedMetaData.MovedAt, metaData.MovedAt.Format(time.RFC3339), "MOVED AT")
	}

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) iPublishTheCollection(collectionID string) error {
	body := fmt.Sprintf(`{"state": %q}`, store.StatePublished)
	err := c.APIFeature.IPatch(fmt.Sprintf("/collection/%s", collectionID), &messages.PickleDocString{MediaType: "application/json", Content: body})
	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) iPublishTheBundle(bundleID string) error {
	body := fmt.Sprintf(`{"state": %q}`, store.StatePublished)
	err := c.APIFeature.IPatch(fmt.Sprintf("/bundle/%s", bundleID), &messages.PickleDocString{MediaType: "application/json", Content: body})
	assert.NoError(c.APIFeature, err)

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) theFollowingPublishedMessageIsSent(table *godog.Table) error {
	expectedMessage, _ := assistdog.NewDefault().ParseMap(table)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if msg, ok := c.msgs[expectedMessage["path"]]; ok {
			assert.True(c.APIFeature, ok, "Could not find message")
			assert.Equal(c.APIFeature, expectedMessage["path"], msg.Path)
			assert.Equal(c.APIFeature, expectedMessage["etag"], msg.Etag)
			assert.Equal(c.APIFeature, expectedMessage["type"], msg.Type)
			assert.Equal(c.APIFeature, expectedMessage["sizeInBytes"], msg.SizeInBytes)
			return c.APIFeature.StepError()
		}

		time.Sleep(50 * time.Millisecond)
	}

	assert.Fail(c.APIFeature, "Could not find kafka message")
	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) kafkaConsumerGroupIsRunning() error {
	c.msgs = make(map[string]files.FilePublished)
	ctx := context.Background()
	cfg, _ := config.Get()
	minRetry := 1 * time.Millisecond
	maxRetry := 5 * time.Millisecond
	cgConfig := &kafka.ConsumerGroupConfig{
		KafkaVersion:      &cfg.KafkaConfig.Version,
		MinBrokersHealthy: &cfg.KafkaConfig.ProducerMinBrokersHealthy,
		Topic:             cfg.KafkaConfig.StaticFilePublishedTopic,
		GroupName:         "testing-stuff",
		BrokerAddrs:       cfg.KafkaConfig.Addr,
		MinRetryPeriod:    &minRetry,
		MaxRetryPeriod:    &maxRetry,
	}
	c.cg, _ = kafka.NewConsumerGroup(ctx, cgConfig)
	err := c.cg.Start()
	assert.NoError(c.APIFeature, err)

	err = c.cg.RegisterHandler(ctx, func(ctx context.Context, workerID int, msg kafka.Message) error {
		schema := &avro.Schema{
			Definition: `{
					"type": "record",
					"name": "file-published",
					"fields": [
					  {"name": "path", "type": "string"},
					  {"name": "etag", "type": "string"},
					  {"name": "type", "type": "string"},
					  {"name": "sizeInBytes", "type": "string"}
					]
				  }`,
		}
		fp := files.FilePublished{}
		err := schema.Unmarshal(msg.GetData(), &fp)
		assert.NoError(c.APIFeature, err)

		c.msgs[fp.Path] = fp

		return nil
	})
	assert.NoError(c.APIFeature, err)

	for {
		if c.cg.State().String() == "Consuming" {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) iAmInWebMode() error {
	c.isPublishing = false

	return c.APIFeature.StepError()
}

func (c *FilesAPIComponent) iGetFilesInTheCollection(collectionID string) error {
	return c.APIFeature.IGet(fmt.Sprintf("/files?collection_id=%s", collectionID))
}

func (c *FilesAPIComponent) iSetTheBundleIDToForFile(bundleID, path string) error {
	json := fmt.Sprintf(`{"bundle_id": %q}`, bundleID)
	return c.APIFeature.IPatch(fmt.Sprintf("/files/%s", path), &messages.PickleDocString{Content: json})
}

func (c *FilesAPIComponent) iGetFilesInTheBundle(bundleID string) error {
	return c.APIFeature.IGet(fmt.Sprintf("/files?bundle_id=%s", bundleID))
}

func (c *FilesAPIComponent) iGetFilesWithBothCollectionAndBundleID(collectionID, bundleID string) error {
	return c.APIFeature.IGet(fmt.Sprintf("/files?collection_id=%s&bundle_id=%s", collectionID, bundleID))
}

func (c *FilesAPIComponent) iUpdateTheContentItemOfTheFileWith(path string, payload *godog.DocString) error {
	return c.APIFeature.IPut(fmt.Sprintf("/files/%s", path), payload)
}
