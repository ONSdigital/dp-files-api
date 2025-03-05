package store_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StoreSuite struct {
	suite.Suite
	logInterceptor       LogInterceptor
	defaultCollectionID  string
	path                 string
	defaultContext       context.Context
	defaultClock         steps.TestClock
	defaultKafkaProducer kafkatest.IProducerMock
}

var (
	mu sync.Mutex
)

type CollectionCountFunc func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error)
type CollectionFindFunc func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error)
type CollectionFindCursorFunc func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (mongodriver.Cursor, error)
type CollectionFindOneFunc func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error
type CollectionUpdateFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)
type CollectionUpdateManyFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)
type CollectionInsertFunc func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error)
type KafkaSendFunc func(schema *avro.Schema, event interface{}) error

func CollectionFindReturnsValueAndError(value int, expectedError error) CollectionFindFunc {
	return func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
		return value, expectedError
	}
}

func CollectionFindOneSetsResultAndReturnsNil(metadataBytes []byte) CollectionFindOneFunc {
	return func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
		bson.Unmarshal(metadataBytes, result)
		return nil
	}
}
func CollectionFindOneSucceeds() CollectionFindOneFunc {
	metadata := files.StoredRegisteredMetaData{}
	metadataBytes, _ := bson.Marshal(metadata)

	return CollectionFindOneSetsResultAndReturnsNil(metadataBytes)
}

func CollectionFindOneReturnsError(expectedError error) CollectionFindOneFunc {
	return func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
		return expectedError
	}
}

type CollectionFindOneFuncChainEntry struct {
	fun   CollectionFindOneFunc
	times int
}

func CollectionFindOneChain(chain []CollectionFindOneFuncChainEntry) CollectionFindOneFunc {
	currentRun := 0
	return func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
		currentRun++
		run := 0
		for _, item := range chain {
			run += item.times
			if currentRun <= run {
				return item.fun(ctx, filter, result, opts...)
			}
		}
		return errors.New("unexpected CollectionFindOne call: no functions left in the chain")
	}
}

func CollectionFindCursorReturnsCursorAndError(cursor mongodriver.Cursor, expectedError error) CollectionFindCursorFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (mongodriver.Cursor, error) {
		return cursor, expectedError
	}
}

func CollectionUpdateReturnsNilAndNil() CollectionUpdateFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, nil
	}
}

func CollectionUpdateReturnsNilAndError(expectedError error) CollectionUpdateFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, expectedError
	}
}

func CollectionCountReturnsValueAndError(value int, expectedError error) CollectionCountFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
		return value, expectedError
	}
}

func CollectionCountReturnsValueAndNil(value int) CollectionCountFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
		return value, nil
	}
}

func CollectionCountReturnsOneNilWhenFilterContainsAndOrZeroNilWithout() CollectionCountFunc {
	return func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
		bsonFilter := filter.(primitive.M)

		if bsonFilter["$and"] == nil {
			// Count of all files in collection
			return 1, nil
		}

		return 0, nil
	}
}

func CollectionUpdateManyReturnsNilAndError(expectedError error) CollectionUpdateManyFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, expectedError
	}
}

func CollectionUpdateManyReturnsNilAndNil() CollectionUpdateManyFunc {
	return func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
		return nil, nil
	}
}

func CollectionFindSetsResultsReturnsValueAndNil(metadataBytes []byte, value int) CollectionFindFunc {
	return func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
		result := files.StoredRegisteredMetaData{}
		bson.Unmarshal(metadataBytes, &result)

		resultPointer := results.(*[]files.StoredRegisteredMetaData)
		*resultPointer = []files.StoredRegisteredMetaData{result}

		return value, nil
	}
}

func CollectionFindReturnsMetadataOnFilter(metadata []files.StoredRegisteredMetaData, expectedFilter interface{}) CollectionFindFunc {
	return func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
		if !reflect.DeepEqual(filter, expectedFilter) {
			return 0, fmt.Errorf("filter %#v doesn't match expected %#v", filter, expectedFilter)
		}

		resultPointer := results.(*[]files.StoredRegisteredMetaData)
		*resultPointer = metadata
		return len(metadata), nil
	}
}

func CollectionInsertReturnsNilAndError(expectedError error) CollectionInsertFunc {
	return func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
		return nil, expectedError
	}
}

func CollectionInsertReturnsNilAndNil() CollectionInsertFunc {
	return func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
		return nil, nil
	}
}

func CursorReturnsNumberOfNext(number int) func(ctx context.Context) bool {
	return func(ctx context.Context) bool {
		mu.Lock()
		defer mu.Unlock()
		if number > 0 {
			number--
			return true
		}
		return false
	}
}

func KafkaSendReturnsError(expectedError error) KafkaSendFunc {
	return func(schema *avro.Schema, event interface{}) error {
		return expectedError
	}
}

func KafkaSendReturnsNil() KafkaSendFunc {
	return func(schema *avro.Schema, event interface{}) error {
		return nil
	}
}

func (suite *StoreSuite) SetupTest() {
	suite.defaultCollectionID = "123456"
	suite.path = "test.txt"
	suite.defaultContext = context.Background()
	suite.defaultClock = steps.TestClock{}
	suite.defaultKafkaProducer = kafkatest.IProducerMock{}
	suite.logInterceptor = NewLogInterceptor()
}

func (suite *StoreSuite) etagReference(metadata files.StoredRegisteredMetaData) files.FileEtagChange {
	return files.FileEtagChange{
		Path: metadata.Path,
		Etag: metadata.Etag,
	}
}

func (suite *StoreSuite) assertImmutableFieldsUnchanged(metadata, actualMetadata files.StoredRegisteredMetaData) {
	suite.Equal(metadata.Path, actualMetadata.Path)
	suite.Equal(metadata.IsPublishable, actualMetadata.IsPublishable)
	suite.Equal(metadata.CollectionID, actualMetadata.CollectionID)
	suite.Equal(metadata.Title, actualMetadata.Title)
	suite.Equal(metadata.SizeInBytes, actualMetadata.SizeInBytes)
	suite.Equal(metadata.Type, actualMetadata.Type)
	suite.Equal(metadata.Licence, actualMetadata.Licence)
	suite.Equal(metadata.LicenceURL, actualMetadata.LicenceURL)
	suite.Equal(metadata.Etag, actualMetadata.Etag)
}

func (suite *StoreSuite) generateTestTime(addedDuration time.Duration) time.Time {
	return time.Now().Add(time.Second * addedDuration).Round(time.Second).UTC()
}

func (suite *StoreSuite) generateMetadata(collectionID string) files.StoredRegisteredMetaData {
	createdAt := suite.generateTestTime(1)
	lastModified := suite.generateTestTime(2)
	uploadCompletedAt := suite.generateTestTime(3)
	publishedAt := suite.generateTestTime(4)
	movedAt := suite.generateTestTime(5)

	return files.StoredRegisteredMetaData{
		Path:              suite.path,
		IsPublishable:     true,
		CollectionID:      &collectionID,
		Title:             "Test file",
		SizeInBytes:       10,
		Type:              "text/plain",
		Licence:           "MIT",
		LicenceURL:        "https://opensource.org/licenses/MIT",
		CreatedAt:         createdAt,
		LastModified:      lastModified,
		UploadCompletedAt: &uploadCompletedAt,
		PublishedAt:       &publishedAt,
		MovedAt:           &movedAt,
		State:             store.StateMoved,
		Etag:              "1234567",
	}
}

func (suite *StoreSuite) generatePublishedCollectionInfo(collectionID string) files.StoredCollection {
	lastModified := suite.generateTestTime(10)
	publishedAt := suite.generateTestTime(11)

	return files.StoredCollection{
		ID:           collectionID,
		LastModified: lastModified,
		PublishedAt:  &publishedAt,
		State:        store.StatePublished,
	}
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

type LogInterceptor struct {
	logBuffer                     *bytes.Buffer
	defaultLogDestination         io.Writer
	defaultFallbackLogDestination io.Writer
}

func (l *LogInterceptor) Start() {
	log.SetDestination(l.logBuffer, l.logBuffer)
}
func (l *LogInterceptor) Stop() {
	l.logBuffer.Reset()
	log.SetDestination(l.defaultLogDestination, l.defaultFallbackLogDestination)
}

func (l *LogInterceptor) GetLogEvent() string {
	logResult, _ := io.ReadAll(l.logBuffer)
	logOut := make(map[string]interface{})
	json.Unmarshal(logResult, &logOut)

	return logOut["event"].(string)
}

func (l *LogInterceptor) GetLogEvents(eventName string) map[int]map[string]interface{} {
	retVal := make(map[int]map[string]interface{})
	logResult, _ := io.ReadAll(l.logBuffer)
	logz := strings.Split(string(logResult), "\n")
	counter := 0
	for _, line := range logz {
		logOut := make(map[string]interface{})
		json.Unmarshal([]byte(line), &logOut)
		evt, ok := logOut["event"]
		if ok && evt.(string) == eventName {
			retVal[counter] = logOut["data"].(map[string]interface{})
			counter++
		}
	}

	return retVal
}

func (l *LogInterceptor) IsEventPresent(eventName string) bool {
	logResult, _ := io.ReadAll(l.logBuffer)
	logz := strings.Split(string(logResult), "\n")
	for _, line := range logz {
		logOut := make(map[string]interface{})
		json.Unmarshal([]byte(line), &logOut)
		evt, ok := logOut["event"]
		if ok && evt.(string) == eventName {
			return true
		}
	}

	return false
}

func NewLogInterceptor() LogInterceptor {
	return LogInterceptor{
		logBuffer:                     &bytes.Buffer{},
		defaultLogDestination:         os.Stdout,
		defaultFallbackLogDestination: os.Stderr,
	}
}
