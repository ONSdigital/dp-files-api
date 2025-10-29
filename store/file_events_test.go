package store_test

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo/mock"
	"github.com/ONSdigital/dp-files-api/sdk"
	"github.com/ONSdigital/dp-files-api/store"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (suite *StoreSuite) TestCreateFileEventSuccess() {
	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			return &mongodriver.CollectionInsertResult{}, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
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

	event := &sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123"},
		Action:      sdk.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &sdk.FileMetaData{Path: "file.csv", Type: "csv"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Equal("failed to insert file event", logEvent)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
}

func (suite *StoreSuite) TestCreateFileEventSetsCreatedAtTimestamp() {
	var capturedEvent *sdk.FileEvent

	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			capturedEvent = document.(*sdk.FileEvent)
			return &mongodriver.CollectionInsertResult{}, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user123", Email: "user@example.com"},
		Action:      sdk.ActionCreate,
		Resource:    "/files/test.csv",
		File:        &sdk.FileMetaData{Path: "test.csv", Type: "text/csv"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	suite.NoError(err)
	suite.NotNil(capturedEvent.CreatedAt)
	suite.Equal(suite.defaultClock.GetCurrentTime(), *capturedEvent.CreatedAt)
}

func (suite *StoreSuite) TestCreateFileEventPreservesEventData() {
	var capturedEvent *sdk.FileEvent

	fileEventsCollection := mock.MongoCollectionMock{
		InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
			capturedEvent = document.(*sdk.FileEvent)
			return &mongodriver.CollectionInsertResult{}, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	event := &sdk.FileEvent{
		RequestedBy: &sdk.RequestedBy{ID: "user456", Email: "test@example.com"},
		Action:      sdk.ActionDelete,
		Resource:    "/files/old-file.xls",
		File:        &sdk.FileMetaData{Path: "old-file.xls", Type: "application/xls"},
	}

	err := subject.CreateFileEvent(suite.defaultContext, event)

	suite.NoError(err)
	suite.Equal("user456", capturedEvent.RequestedBy.ID)
	suite.Equal("test@example.com", capturedEvent.RequestedBy.Email)
	suite.Equal(sdk.ActionDelete, capturedEvent.Action)
	suite.Equal("/files/old-file.xls", capturedEvent.Resource)
	suite.Equal("old-file.xls", capturedEvent.File.Path)
}

func (suite *StoreSuite) TestGetFileEventsSuccess() {
	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 5, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			events := results.(*[]sdk.FileEvent)
			*events = []sdk.FileEvent{
				{
					RequestedBy: &sdk.RequestedBy{ID: "user123"},
					Action:      sdk.ActionRead,
					Resource:    "/downloads/file.csv",
					File:        &sdk.FileMetaData{Path: "file.csv"},
				},
			}
			return 1, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "", nil, nil)

	suite.NoError(err)
	suite.NotNil(eventsList)
	suite.Equal(1, eventsList.Count)
	suite.Equal(20, eventsList.Limit)
	suite.Equal(0, eventsList.Offset)
	suite.Equal(5, eventsList.TotalCount)
	suite.Equal(1, len(eventsList.Items))
}

func (suite *StoreSuite) TestGetFileEventsWithPathFilter() {
	var capturedFilter interface{}

	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 2, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			capturedFilter = filter
			events := results.(*[]sdk.FileEvent)
			*events = []sdk.FileEvent{}
			return 0, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "test-file.csv", nil, nil)

	suite.NoError(err)
	suite.NotNil(eventsList)

	filterMap := capturedFilter.(bson.M)
	suite.Equal("test-file.csv", filterMap["file.path"])
}

func (suite *StoreSuite) TestGetFileEventsWithDateFilters() {
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	var capturedFilter interface{}

	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 3, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			capturedFilter = filter
			events := results.(*[]sdk.FileEvent)
			*events = []sdk.FileEvent{}
			return 0, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "", &after, &before)

	suite.NoError(err)
	suite.NotNil(eventsList)

	filterMap := capturedFilter.(bson.M)
	hasOrClause := filterMap["$or"] != nil
	hasDirectCreatedAt := filterMap["created_at"] != nil || filterMap["createdat"] != nil
	suite.True(hasOrClause || hasDirectCreatedAt, "Filter should contain date range either in $or clause or directly")
}

func (suite *StoreSuite) TestGetFileEventsWithPagination() {
	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 100, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			suite.Equal(3, len(opts))
			events := results.(*[]sdk.FileEvent)
			*events = []sdk.FileEvent{}
			return 0, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 10, 50, "", nil, nil)

	suite.NoError(err)
	suite.NotNil(eventsList)
	suite.Equal(10, eventsList.Limit)
	suite.Equal(50, eventsList.Offset)
	suite.Equal(100, eventsList.TotalCount)
}

func (suite *StoreSuite) TestGetFileEventsPathNotFound() {
	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "nonexistent.csv", nil, nil)

	suite.Nil(eventsList)
	suite.Error(err)
	suite.ErrorIs(err, store.ErrPathNotFound)
}

func (suite *StoreSuite) TestGetFileEventsCountError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("database connection error")

	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, expectedError
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "", nil, nil)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Nil(eventsList)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.Equal("failed to count file events", logEvent)
}

func (suite *StoreSuite) TestGetFileEventsFindError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("query execution error")

	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 10, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, expectedError
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "", nil, nil)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Nil(eventsList)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.Equal("failed to find file events", logEvent)
}

func (suite *StoreSuite) TestGetFileEventsWithAllFilters() {
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 5, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			events := results.(*[]sdk.FileEvent)
			*events = []sdk.FileEvent{
				{
					RequestedBy: &sdk.RequestedBy{ID: "user123"},
					Action:      sdk.ActionRead,
					Resource:    "/downloads/data.csv",
					File:        &sdk.FileMetaData{Path: "data.csv"},
				},
			}
			return 1, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 50, 10, "data.csv", &after, &before)

	suite.NoError(err)
	suite.NotNil(eventsList)
	suite.Equal(1, eventsList.Count)
	suite.Equal(50, eventsList.Limit)
	suite.Equal(10, eventsList.Offset)
	suite.Equal(5, eventsList.TotalCount)
	suite.Equal(1, len(eventsList.Items))
	suite.Equal("data.csv", eventsList.Items[0].File.Path)
}

func (suite *StoreSuite) TestGetFileEventsEmptyResult() {
	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, nil
		},
		FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
			events := results.(*[]sdk.FileEvent)
			*events = []sdk.FileEvent{}
			return 0, nil
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "", nil, nil)

	suite.NoError(err)
	suite.NotNil(eventsList)
	suite.Equal(0, eventsList.Count)
	suite.Equal(0, eventsList.TotalCount)
	suite.Equal(0, len(eventsList.Items))
}

func (suite *StoreSuite) TestGetFileEventsPathCheckError() {
	suite.logInterceptor.Start()
	defer suite.logInterceptor.Stop()

	expectedError := errors.New("database error during path check")

	fileEventsCollection := mock.MongoCollectionMock{
		CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
			return 0, expectedError
		},
	}

	cfg, _ := config.Get()
	subject := store.NewStore(nil, nil, nil, &fileEventsCollection, nil, suite.defaultClock, nil, cfg)

	eventsList, err := subject.GetFileEvents(suite.defaultContext, 20, 0, "test-file.csv", nil, nil)

	logEvent := suite.logInterceptor.GetLogEvent()

	suite.Nil(eventsList)
	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.Equal("failed to check if path exists", logEvent)
}
