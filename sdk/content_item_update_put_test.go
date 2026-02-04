package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContentItemUpdatePut_success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		responseBody, err := json.Marshal(exampleStoredRegisteredMetaData)
		So(err, ShouldBeNil)

		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(responseBody)))}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When UpdateContentItem is called with a leading slash", func() {
			contentItem := api.ContentItem{
				DatasetID: "test_dataset_id",
				Edition:   "test_edition",
				Version:   "2",
			}
			metadata, err := client.UpdateContentItem(context.Background(), "/path/to/file.txt", contentItem, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected updated contentItem in the metadata is returned", func() {
				So(metadata.ContentItem.DatasetID, ShouldEqual, "test_dataset_id")
				So(metadata.ContentItem.Edition, ShouldEqual, "test_edition")
				So(metadata.ContentItem.Version, ShouldEqual, "2")
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPut)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})

		Convey("When UpdateContentItem is called without a leading slash", func() {
			contentItem := api.ContentItem{
				DatasetID: "test_dataset_id",
				Edition:   "test_edition",
				Version:   "2",
			}
			metadata, err := client.UpdateContentItem(context.Background(), "/path/to/file.txt", contentItem, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected updated contentItem in the metadata is returned", func() {
				So(metadata.ContentItem.DatasetID, ShouldEqual, "test_dataset_id")
				So(metadata.ContentItem.Edition, ShouldEqual, "test_edition")
				So(metadata.ContentItem.Version, ShouldEqual, "2")
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPut)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})
	})
}

func TestContentItemUpdatePut_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that fails when Do() is called", t, func() {
		mockClienter := newMockClienter(nil, errExpectedDoFailure)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When UpdateContentItem is called", func() {
			contentItem := api.ContentItem{
				DatasetID: "test_dataset_id",
				Edition:   "test_edition",
				Version:   "2",
			}
			metadata, err := client.UpdateContentItem(context.Background(), "/path/to/file.txt", contentItem, testHeaders)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, errExpectedDoFailure)
			})

			Convey("And no metadata is returned", func() {
				So(metadata, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})

	Convey("Given a files-api client that returns an unexpected status code", t, func() {
		body := `{"errors":[{"errorCode":"FileNotRegistered","description":"file not registered"}]}`
		mockClienter := newMockClienter(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When UpdateContentItem is called", func() {
			contentItem := api.ContentItem{
				DatasetID: "test_dataset_id",
				Edition:   "test_edition",
				Version:   "2",
			}
			metadata, err := client.UpdateContentItem(context.Background(), "/path/to/file.txt", contentItem, testHeaders)

			Convey("Then the expected API error is returned", func() {
				expectedError := &APIError{
					StatusCode: http.StatusNotFound,
					Errors: &api.JSONErrors{
						Error: []api.JSONError{
							{
								Code:        "FileNotRegistered",
								Description: "file not registered",
							},
						},
					},
				}
				So(err, ShouldResemble, expectedError)
				So(err.Error(), ShouldEqual, expectedError.Error())
			})

			Convey("And no metadata is returned", func() {
				So(metadata, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})
}
