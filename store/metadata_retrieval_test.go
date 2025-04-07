package store_test

import (
	"errors"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestGetFileMetadataNotFoundError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	_, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("file metadata not found", logEvent)
	suite.Equal(store.ErrFileNotRegistered, err)
}

func (suite *StoreSuite) TestGetFileMetadataOtherError() {
	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(errors.New("find error")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	_, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.EqualError(err, "find error")
}

func (suite *StoreSuite) TestGetFileMetadataNoCollectionPatching() {
	expectedMetadata := suite.generateCollectionMetadata(suite.defaultCollectionID)

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	actualMetadata, _ := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFileMetadataCollectionError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedMetadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
	expectedMetadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(expectedMetadata)

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(errors.New("collection error")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	actualMetadata, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection metadata fetch error", logEvent)
	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFileMetadataWithCollectionPatching() {
	metadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(metadata)

	collection := suite.generatePublishedCollectionInfo(suite.defaultCollectionID)
	collectionBytes, _ := bson.Marshal(collection)

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(collectionBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	actualMetadata, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.NoError(err)
	suite.Equal(collection.State, actualMetadata.State)
	suite.NotEqual(metadata.State, actualMetadata.State)
	suite.Equal(collection.PublishedAt, actualMetadata.PublishedAt)
	suite.NotEqual(metadata.PublishedAt, actualMetadata.PublishedAt)
	suite.Equal(collection.LastModified, actualMetadata.LastModified)
	suite.NotEqual(metadata.LastModified, actualMetadata.LastModified)
}

func (suite *StoreSuite) TestGetFileMetadataWithBundleID() {
	bundleID := "test-bundle-id"
	expectedMetadata := files.StoredRegisteredMetaData{
		Path:     suite.path,
		State:    store.StateUploaded,
		BundleID: &bundleID,
	}

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	publishedAt := suite.generateTestTime(1)
	lastModified := suite.generateTestTime(2)
	bundle := files.StoredBundle{
		ID:           bundleID,
		State:        store.StatePublished,
		PublishedAt:  &publishedAt,
		LastModified: lastModified,
	}
	bundleBytes, _ := bson.Marshal(bundle)

	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.NoError(err)
	suite.Equal(store.StatePublished, actualMetadata.State)
	suite.Equal(publishedAt, *actualMetadata.PublishedAt)
	suite.Equal(lastModified, actualMetadata.LastModified)
}

func (suite *StoreSuite) TestGetFileMetadataWithBundleError() {
	bundleID := "test-bundle-id"
	expectedMetadata := files.StoredRegisteredMetaData{
		Path:     suite.path,
		State:    store.StateUploaded,
		BundleID: &bundleID,
	}

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(errors.New("bundle error")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.NoError(err)
	suite.Equal(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoPatching() {
	metadata1 := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadata1.Path += "1"
	metadata2 := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadata2.Path += "2"

	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{metadata1, metadata2},
			bson.M{"collection_id": suite.defaultCollectionID},
		),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(nil),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata1, metadata2}
	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID, "")

	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataWithPatching() {
	metadata1 := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadata1.Path += "1"
	metadata1.State = store.StateUploaded
	metadata2 := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadata2.Path += "2"
	metadata2.State = store.StateUploaded

	collection := suite.generatePublishedCollectionInfo(suite.defaultCollectionID)
	collectionBytes, _ := bson.Marshal(collection)

	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{metadata1, metadata2},
			bson.M{"collection_id": suite.defaultCollectionID},
		),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(collectionBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata1, metadata2}
	expectedMetadata[0].State = store.StatePublished
	expectedMetadata[0].PublishedAt = collection.PublishedAt
	expectedMetadata[0].LastModified = collection.LastModified
	expectedMetadata[1].State = store.StatePublished
	expectedMetadata[1].PublishedAt = collection.PublishedAt
	expectedMetadata[1].LastModified = collection.LastModified

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID, "")

	suite.NoError(err)
	suite.NotEqual(metadata1.State, collection.State)
	suite.NotEqual(metadata1.PublishedAt, collection.PublishedAt)
	suite.NotEqual(metadata1.LastModified, collection.LastModified)
	suite.NotEqual(metadata2.State, collection.State)
	suite.NotEqual(metadata2.PublishedAt, collection.PublishedAt)
	suite.NotEqual(metadata2.LastModified, collection.LastModified)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataNoResult() {
	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{},
			bson.M{"collection_id": "INVALID_COLLECTION_ID"},
		),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(nil),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "INVALID_COLLECTION_ID", "")

	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataFindError() {
	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsValueAndError(0, errors.New("find error")),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(nil),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID, "")

	suite.EqualError(err, "find error")
	suite.Nil(actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataCollectionError() {
	metadata := suite.generateCollectionMetadata(suite.defaultCollectionID)
	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{metadata},
			bson.M{"collection_id": suite.defaultCollectionID},
		),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(errors.New("collection error")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID, "")

	suite.NoError(err)
	suite.Exactly([]files.StoredRegisteredMetaData{metadata}, actualMetadata)
}

// New tests for GetFilesMetadata with Bundle ID
func (suite *StoreSuite) TestGetFilesMetadataWithBundleID() {
	metadata1 := suite.generateBundleMetadata(suite.defaultBundleID)
	metadata1.Path += "1"
	metadata1.State = store.StateUploaded
	metadata2 := suite.generateBundleMetadata(suite.defaultBundleID)
	metadata2.Path += "2"
	metadata2.State = store.StateUploaded

	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{metadata1, metadata2},
			bson.M{"bundle_id": suite.defaultBundleID},
		),
	}
	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(nil),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata1, metadata2}
	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "", suite.defaultBundleID)

	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataWithBundlePatching() {
	metadata1 := suite.generateBundleMetadata(suite.defaultBundleID)
	metadata1.Path += "1"
	metadata1.State = store.StateUploaded
	metadata2 := suite.generateBundleMetadata(suite.defaultBundleID)
	metadata2.Path += "2"
	metadata2.State = store.StateUploaded

	bundle := suite.generatePublishedBundleInfo(suite.defaultBundleID)
	bundleBytes, _ := bson.Marshal(bundle)

	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{metadata1, metadata2},
			bson.M{"bundle_id": suite.defaultBundleID},
		),
	}
	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata1, metadata2}
	expectedMetadata[0].State = store.StatePublished
	expectedMetadata[0].PublishedAt = bundle.PublishedAt
	expectedMetadata[1].State = store.StatePublished
	expectedMetadata[1].PublishedAt = bundle.PublishedAt

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "", suite.defaultBundleID)

	suite.NoError(err)
	suite.NotEqual(metadata1.State, bundle.State)
	suite.NotEqual(metadata1.PublishedAt, bundle.PublishedAt)
	suite.NotEqual(metadata2.State, bundle.State)
	suite.NotEqual(metadata2.PublishedAt, bundle.PublishedAt)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataWithBundleError() {
	metadata := suite.generateBundleMetadata(suite.defaultBundleID)
	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{metadata},
			bson.M{"bundle_id": suite.defaultBundleID},
		),
	}
	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(errors.New("bundle error")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "", suite.defaultBundleID)

	suite.NoError(err)
	suite.Exactly([]files.StoredRegisteredMetaData{metadata}, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataBundleFindError() {
	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsValueAndError(0, errors.New("find error")),
	}
	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(nil),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "", suite.defaultBundleID)

	suite.EqualError(err, "find error")
	suite.Nil(actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataBundleNoResult() {
	metadataColl := mock.MongoCollectionMock{
		FindFunc: CollectionFindReturnsMetadataOnFilter(
			[]files.StoredRegisteredMetaData{},
			bson.M{"bundle_id": "INVALID_BUNDLE_ID"},
		),
	}
	bundleColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(nil),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, nil, &bundleColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "", "INVALID_BUNDLE_ID")

	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestPatchMetadataNilMetadata() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	var metadata *files.StoredRegisteredMetaData
	collection := &files.StoredCollection{}

	subject.PatchFilePublishMetadata(metadata, collection)
	suite.Nil(metadata)
}

func (suite *StoreSuite) TestPatchMetadataNilCollection() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	collectionID := "coll1"
	metadata := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StateUploaded,
		CollectionID: &collectionID,
	}
	metadataExpected := metadata

	subject.PatchFilePublishMetadata(&metadata, nil)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchMetadataNilCollectionID() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	metadata := files.StoredRegisteredMetaData{
		Path:  "path1",
		State: store.StateUploaded,
	}
	collection := &files.StoredCollection{
		ID:    "coll1",
		State: store.StatePublished,
	}

	metadataExpected := metadata

	subject.PatchFilePublishMetadata(&metadata, collection)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchMetadataCollectionIDMismatch() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	collectionID := "coll1"
	metadata := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StateUploaded,
		CollectionID: &collectionID,
	}
	collection := &files.StoredCollection{
		ID:    "a-different-collection-id",
		State: store.StatePublished,
	}

	metadataExpected := metadata

	subject.PatchFilePublishMetadata(&metadata, collection)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchMetadataBadState() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	collectionID := "coll1"
	metadata := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StateMoved,
		CollectionID: &collectionID,
	}
	collection := &files.StoredCollection{
		ID:    collectionID,
		State: store.StatePublished,
	}

	metadataExpected := metadata

	subject.PatchFilePublishMetadata(&metadata, collection)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchMetadataBadCollectionState() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	collectionID := "coll1"
	metadata := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StateUploaded,
		CollectionID: &collectionID,
	}
	collection := &files.StoredCollection{
		ID: collectionID,
	}

	metadataExpected := metadata

	subject.PatchFilePublishMetadata(&metadata, collection)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchMetadataSuccess() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	collectionID := "coll1"
	publishedAt := suite.generateTestTime(1)
	lastModified := suite.generateTestTime(2)
	metadata := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StateUploaded,
		CollectionID: &collectionID,
	}
	collection := &files.StoredCollection{
		ID:           collectionID,
		State:        store.StatePublished,
		PublishedAt:  &publishedAt,
		LastModified: lastModified,
	}

	metadataExpected := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StatePublished,
		CollectionID: &collectionID,
		PublishedAt:  &publishedAt,
		LastModified: lastModified,
	}

	subject.PatchFilePublishMetadata(&metadata, collection)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataNilMetadata() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	var metadata *files.StoredRegisteredMetaData
	bundle := &files.StoredBundle{}

	subject.PatchFilePublishBundleMetadata(metadata, bundle)
	suite.Nil(metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataNilBundle() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	bundleID := "bundle1"
	metadata := files.StoredRegisteredMetaData{
		Path:     "path1",
		State:    store.StateUploaded,
		BundleID: &bundleID,
	}
	metadataExpected := metadata

	subject.PatchFilePublishBundleMetadata(&metadata, nil)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataNilBundleID() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	metadata := files.StoredRegisteredMetaData{
		Path:  "path1",
		State: store.StateUploaded,
	}
	bundle := &files.StoredBundle{
		ID:    "bundle1",
		State: store.StatePublished,
	}

	metadataExpected := metadata

	subject.PatchFilePublishBundleMetadata(&metadata, bundle)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataBundleIDMismatch() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	bundleID := "bundle1"
	metadata := files.StoredRegisteredMetaData{
		Path:     "path1",
		State:    store.StateUploaded,
		BundleID: &bundleID,
	}
	bundle := &files.StoredBundle{
		ID:    "a-different-bundle-id",
		State: store.StatePublished,
	}

	metadataExpected := metadata

	subject.PatchFilePublishBundleMetadata(&metadata, bundle)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataBadState() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	bundleID := "bundle1"
	metadata := files.StoredRegisteredMetaData{
		Path:     "path1",
		State:    store.StateMoved,
		BundleID: &bundleID,
	}
	bundle := &files.StoredBundle{
		ID:    bundleID,
		State: store.StatePublished,
	}

	metadataExpected := metadata

	subject.PatchFilePublishBundleMetadata(&metadata, bundle)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataBadBundleState() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	bundleID := "bundle1"
	metadata := files.StoredRegisteredMetaData{
		Path:     "path1",
		State:    store.StateUploaded,
		BundleID: &bundleID,
	}
	bundle := &files.StoredBundle{
		ID: bundleID,
	}

	metadataExpected := metadata

	subject.PatchFilePublishBundleMetadata(&metadata, bundle)
	suite.Exactly(metadataExpected, metadata)
}

func (suite *StoreSuite) TestPatchBundleMetadataSuccess() {
	subject := store.NewStore(nil, nil, nil, nil, suite.defaultClock, nil, nil)

	bundleID := "bundle1"
	publishedAt := suite.generateTestTime(1)
	lastModified := suite.generateTestTime(2)
	metadata := files.StoredRegisteredMetaData{
		Path:     "path1",
		State:    store.StateUploaded,
		BundleID: &bundleID,
	}
	bundle := &files.StoredBundle{
		ID:           bundleID,
		State:        store.StatePublished,
		PublishedAt:  &publishedAt,
		LastModified: lastModified,
	}

	metadataExpected := files.StoredRegisteredMetaData{
		Path:         "path1",
		State:        store.StatePublished,
		BundleID:     &bundleID,
		PublishedAt:  &publishedAt,
		LastModified: lastModified,
	}

	subject.PatchFilePublishBundleMetadata(&metadata, bundle)
	suite.Exactly(metadataExpected, metadata)
}
