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
	subject := store.NewStore(&collection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
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
	subject := store.NewStore(&collection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	_, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.EqualError(err, "find error")
}

func (suite *StoreSuite) TestGetFileMetadataNoCollectionPatching() {
	expectedMetadata := suite.generateMetadata(suite.defaultCollectionID)

	metadataBytes, _ := bson.Marshal(expectedMetadata)

	collection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&collection, nil, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	actualMetadata, _ := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFileMetadataCollectionError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedMetadata := suite.generateMetadata(suite.defaultCollectionID)
	expectedMetadata.State = store.StateUploaded
	metadataBytes, _ := bson.Marshal(expectedMetadata)

	metadataColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}
	collectionColl := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(errors.New("collection error")),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	actualMetadata, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("collection metadata fetch error", logEvent)
	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFileMetadataWithCollectionPatching() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
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
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)
	actualMetadata, err := subject.GetFileMetadata(suite.defaultContext, suite.path)

	suite.NoError(err)
	suite.Equal(collection.State, actualMetadata.State)
	suite.NotEqual(metadata.State, actualMetadata.State)
	suite.Equal(collection.PublishedAt, actualMetadata.PublishedAt)
	suite.NotEqual(metadata.PublishedAt, actualMetadata.PublishedAt)
	suite.Equal(collection.LastModified, actualMetadata.LastModified)
	suite.NotEqual(metadata.LastModified, actualMetadata.LastModified)
}

func (suite *StoreSuite) TestGetFilesMetadataNoPatching() {
	metadata1 := suite.generateMetadata(suite.defaultCollectionID)
	metadata1.Path += "1"
	metadata2 := suite.generateMetadata(suite.defaultCollectionID)
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
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata1, metadata2}
	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID)

	suite.NoError(err)
	suite.Exactly(expectedMetadata, actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataWithPatching() {
	metadata1 := suite.generateMetadata(suite.defaultCollectionID)
	metadata1.Path += "1"
	metadata1.State = store.StateUploaded
	metadata2 := suite.generateMetadata(suite.defaultCollectionID)
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
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := []files.StoredRegisteredMetaData{metadata1, metadata2}
	expectedMetadata[0].State = store.StatePublished
	expectedMetadata[0].PublishedAt = collection.PublishedAt
	expectedMetadata[0].LastModified = collection.LastModified
	expectedMetadata[1].State = store.StatePublished
	expectedMetadata[1].PublishedAt = collection.PublishedAt
	expectedMetadata[1].LastModified = collection.LastModified

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID)

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
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	expectedMetadata := make([]files.StoredRegisteredMetaData, 0)
	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, "INVALID_COLLECTION_ID")

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
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID)

	suite.EqualError(err, "find error")
	suite.Nil(actualMetadata)
}

func (suite *StoreSuite) TestGetFilesMetadataCollectionError() {
	metadata := suite.generateMetadata(suite.defaultCollectionID)
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
	subject := store.NewStore(&metadataColl, &collectionColl, &suite.defaultKafkaProducer, suite.defaultClock, nil, cfg)

	actualMetadata, err := subject.GetFilesMetadata(suite.defaultContext, suite.defaultCollectionID)

	suite.NoError(err)
	suite.Exactly([]files.StoredRegisteredMetaData{metadata}, actualMetadata)
}

func (suite *StoreSuite) TestPatchMetadataNilMetadata() {
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

	var metadata *files.StoredRegisteredMetaData
	collection := &files.StoredCollection{}

	subject.PatchFilePublishMetadata(metadata, collection)
	suite.Nil(metadata)
}

func (suite *StoreSuite) TestPatchMetadataNilCollection() {
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

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
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

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
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

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
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

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
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

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
	subject := store.NewStore(nil, nil, nil, suite.defaultClock, nil, nil)

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
