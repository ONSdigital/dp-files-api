package store_test

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/store"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

func (suite *StoreSuite) TestCreateFileEventSuccess() {
	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			return &mongodriver.CollectionInsertResult{}, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	suite.NoError(err)
	suite.NotNil(event.CreatedAt)
	suite.Equal(1, len(fileEventsCollection.InsertCalls()))
}

func (suite *StoreSuite) TestCreateFileEventInsertError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("database error")

	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			return nil, expectedError
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to insert file event", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestCreateFileEventSetsCreatedAtTimestamp() {
	var capturedEvent *files.FileEvent

	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			capturedEvent = document.(*files.FileEvent)
			return &mongodriver.CollectionInsertResult{}, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123", Email: "user@example.com"},
		Action:      files.ActionCreate,
		Resource:    "/files/test.csv",
		File:        &files.FileMetaData{Path: "test.csv", Type: "text/csv"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	suite.NoError(err)
	suite.NotNil(capturedEvent.CreatedAt)
	suite.Equal(suite.defaultClock.GetCurrentTime(), *capturedEvent.CreatedAt)
}

func (suite *StoreSuite) TestCreateFileEventPreservesEventData() {
	var capturedEvent *files.FileEvent

	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			capturedEvent = document.(*files.FileEvent)
			return &mongodriver.CollectionInsertResult{}, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user456", Email: "test@example.com"},
		Action:      files.ActionDelete,
		Resource:    "/files/old-file.xls",
		File:        &files.FileMetaData{Path: "old-file.xls", Type: "application/xls"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	suite.NoError(err)
	suite.Equal("user456", capturedEvent.RequestedBy.ID)
	suite.Equal("test@example.com", capturedEvent.RequestedBy.Email)
	suite.Equal(files.ActionDelete, capturedEvent.Action)
	suite.Equal("/files/old-file.xls", capturedEvent.Resource)
	suite.Equal("old-file.xls", capturedEvent.File.Path)
}
