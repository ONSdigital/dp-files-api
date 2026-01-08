package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	exampleFileEvent = files.FileEvent{
		RequestedBy: &files.RequestedBy{ID: "user123"},
		Action:      files.ActionRead,
		Resource:    "/downloads/file.csv",
		File:        &files.FileMetaData{Path: "file.csv", Type: "csv"},
	}
)

func TestCreateFileEvent_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		responseBody, err := json.Marshal(exampleFileEvent)
		So(err, ShouldBeNil)

		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(string(responseBody)))}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When CreateFileEvent is called", func() {
			createdEvent, err := client.CreateFileEvent(context.Background(), exampleFileEvent, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected created event is returned", func() {
				So(createdEvent, ShouldResemble, &exampleFileEvent)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPost)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/file-events")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})
	})
}

func TestCreateFileEvent_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that returns an error on Do()", t, func() {
		mockClienter := newMockClienter(nil, errExpectedDoFailure)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When CreateFileEvent is called", func() {
			createdEvent, err := client.CreateFileEvent(context.Background(), exampleFileEvent, testHeaders)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("failed to execute request: %w", errExpectedDoFailure))
			})

			Convey("And no created event is returned", func() {
				So(createdEvent, ShouldBeNil)
			})
		})
	})

	Convey("Given a files-api client that returns an unexpected status code", t, func() {
		body := `{"errors":[{"errorCode":"InvalidEvent","description":"invalid event data"}]}`
		mockClienter := newMockClienter(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When CreateFileEvent is called", func() {
			createdEvent, err := client.CreateFileEvent(context.Background(), exampleFileEvent, testHeaders)

			Convey("Then the expected API error is returned", func() {
				expectedError := &APIError{
					StatusCode: http.StatusBadRequest,
					Errors: &api.JSONErrors{
						Error: []api.JSONError{
							{
								Code:        "InvalidEvent",
								Description: "invalid event data",
							},
						},
					},
				}
				So(err, ShouldResemble, expectedError)
			})

			Convey("And no created event is returned", func() {
				So(createdEvent, ShouldBeNil)
			})
		})
	})
}
