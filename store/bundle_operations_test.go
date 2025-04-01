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

func (suite *StoreSuite) TestGetBundlePublishedMetadataSuccess() {
	expectedBundle := files.StoredCollection{
		ID:    suite.defaultBundleID,
		State: store.StatePublished,
	}
	bundleBytes, _ := bson.Marshal(expectedBundle)

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	actualBundle, err := subject.GetBundlePublishedMetadata(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.Exactly(expectedBundle, actualBundle)
}

func (suite *StoreSuite) TestGetBundlePublishedMetadataNotFound() {
	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	actualBundle, err := subject.GetBundlePublishedMetadata(suite.defaultContext, "non-existent-bundle-id")

	suite.Error(err)
	suite.ErrorIs(err, store.ErrBundleMetadataNotRegistered)
	suite.Equal(files.StoredCollection{}, actualBundle)
}

func (suite *StoreSuite) TestGetBundlePublishedMetadataUnexpectedError() {
	expectedError := errors.New("unexpected error")

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	actualBundle, err := subject.GetBundlePublishedMetadata(suite.defaultContext, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.Equal(files.StoredCollection{}, actualBundle)
}
func (suite *StoreSuite) TestIsBundlePublishedSuccess() {
	expectedBundle := files.StoredCollection{
		ID:    suite.defaultBundleID,
		State: store.StatePublished,
	}
	bundleBytes, _ := bson.Marshal(expectedBundle)

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.True(isPublished)
}

func (suite *StoreSuite) TestIsBundlePublishedNotPublishedState() {
	expectedBundle := files.StoredCollection{
		ID:    suite.defaultBundleID,
		State: store.StateCreated,
	}
	bundleBytes, _ := bson.Marshal(expectedBundle)

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(bundleBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestIsBundlePublishedFallbackToFileCheckNotPublished() {
	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(store.ErrBundleMetadataNotRegistered),
	}

	metadata := bson.M{
		"state": store.StateCreated,
	}
	metadataBytes, _ := bson.Marshal(metadata)

	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestIsBundlePublishedUnexpectedError() {
	expectedError := errors.New("unexpected error")

	bundlesCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.IsBundlePublished(suite.defaultContext, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestAreAllBundleFilesPublishedNotAllFilesPublished() {
	metadata := files.StoredRegisteredMetaData{
		BundleID: &suite.defaultBundleID,
		State:    store.StateCreated,
	}
	metadataBytes, _ := bson.Marshal(metadata)

	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneSetsResultAndReturnsNil(metadataBytes),
	}

	bundlesCollection := mock.MongoCollectionMock{}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.AreAllBundleFilesPublished(suite.defaultContext, suite.defaultBundleID)

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestAreAllBundleFilesPublishedEmptyBundle() {
	bundlesCollection := mock.MongoCollectionMock{}
	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(mongodriver.ErrNoDocumentFound),
	}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.AreAllBundleFilesPublished(suite.defaultContext, "empty-bundle-id")

	suite.NoError(err)
	suite.False(isPublished)
}

func (suite *StoreSuite) TestAreAllBundleFilesPublishedUnexpectedError() {
	expectedError := errors.New("unexpected error")

	metadataCollection := mock.MongoCollectionMock{
		FindOneFunc: CollectionFindOneReturnsError(expectedError),
	}

	bundlesCollection := mock.MongoCollectionMock{}

	cfg, _ := config.Get()
	subject := store.NewStore(&metadataCollection, nil, &bundlesCollection, nil, suite.defaultClock, nil, cfg)

	isPublished, err := subject.AreAllBundleFilesPublished(suite.defaultContext, suite.defaultBundleID)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.False(isPublished)
}
