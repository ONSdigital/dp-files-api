package sdk

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetFile_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		responseBody, err := json.Marshal(exampleStoredRegisteredMetaData)
		So(err, ShouldBeNil)

		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(responseBody)))}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When GetFile is called with a leading slash", func() {
			metadata, err := client.GetFile(ctx, "/path/to/file.txt")

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected metadata is returned", func() {
				// must check each field individually since StoredRegisteredMetaData does not include time fields in json marshalling
				So(metadata.Path, ShouldEqual, exampleStoredRegisteredMetaData.Path)
				So(metadata.IsPublishable, ShouldEqual, exampleStoredRegisteredMetaData.IsPublishable)
				So(metadata.CollectionID, ShouldResemble, exampleStoredRegisteredMetaData.CollectionID)
				So(metadata.BundleID, ShouldResemble, exampleStoredRegisteredMetaData.BundleID)
				So(metadata.Title, ShouldEqual, exampleStoredRegisteredMetaData.Title)
				So(metadata.SizeInBytes, ShouldEqual, exampleStoredRegisteredMetaData.SizeInBytes)
				So(metadata.Type, ShouldEqual, exampleStoredRegisteredMetaData.Type)
				So(metadata.Licence, ShouldEqual, exampleStoredRegisteredMetaData.Licence)
				So(metadata.LicenceURL, ShouldEqual, exampleStoredRegisteredMetaData.LicenceURL)
				So(metadata.State, ShouldEqual, exampleStoredRegisteredMetaData.State)
				So(metadata.Etag, ShouldEqual, exampleStoredRegisteredMetaData.Etag)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodGet)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})

		Convey("When GetFile is called without a leading slash", func() {
			metadata, err := client.GetFile(ctx, "path/to/file.txt")

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected metadata is returned", func() {
				// must check each field individually since StoredRegisteredMetaData does not include time fields in json marshalling
				So(metadata.Path, ShouldEqual, exampleStoredRegisteredMetaData.Path)
				So(metadata.IsPublishable, ShouldEqual, exampleStoredRegisteredMetaData.IsPublishable)
				So(metadata.CollectionID, ShouldResemble, exampleStoredRegisteredMetaData.CollectionID)
				So(metadata.BundleID, ShouldResemble, exampleStoredRegisteredMetaData.BundleID)
				So(metadata.Title, ShouldEqual, exampleStoredRegisteredMetaData.Title)
				So(metadata.SizeInBytes, ShouldEqual, exampleStoredRegisteredMetaData.SizeInBytes)
				So(metadata.Type, ShouldEqual, exampleStoredRegisteredMetaData.Type)
				So(metadata.Licence, ShouldEqual, exampleStoredRegisteredMetaData.Licence)
				So(metadata.LicenceURL, ShouldEqual, exampleStoredRegisteredMetaData.LicenceURL)
				So(metadata.State, ShouldEqual, exampleStoredRegisteredMetaData.State)
				So(metadata.Etag, ShouldEqual, exampleStoredRegisteredMetaData.Etag)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodGet)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})
	})
}

func TestGetFile_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that fails when Do() is called", t, func() {
		mockClienter := newMockClienter(nil, expectedDoErr)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When GetFile is called", func() {
			metadata, err := client.GetFile(ctx, "/path/to/file.txt")

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, expectedDoErr)
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

		Convey("When GetFile is called", func() {
			metadata, err := client.GetFile(ctx, "/path/to/file.txt")

			Convey("Then the expected API error is returned", func() {
				expectedError := &APIError{
					StatusCode: http.StatusNotFound,
					Errors: &api.JsonErrors{
						Error: []api.JsonError{
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
