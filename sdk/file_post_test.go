package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRegisterFile_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusCreated}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When RegisterFile is called", func() {
			isPublishable := true
			collectionID := exampleCollectionID
			metadata := files.StoredRegisteredMetaData{
				Path:          "path/to/file.txt",
				IsPublishable: isPublishable,
				CollectionID:  &collectionID,
				Title:         "Test File",
				SizeInBytes:   12345,
				Type:          "text/plain",
				Licence:       "OGL v3",
				LicenceURL:    "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
			}

			err := client.RegisterFile(context.Background(), metadata, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPost)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
				So(actualCall.Req.Header.Get("Content-Type"), ShouldEqual, "application/json")

				bodyBytes, err := io.ReadAll(actualCall.Req.Body)
				So(err, ShouldBeNil)

				var actualMetadata files.StoredRegisteredMetaData
				err = json.Unmarshal(bodyBytes, &actualMetadata)
				So(err, ShouldBeNil)
				So(actualMetadata, ShouldResemble, metadata)
			})
		})
	})
}

func TestRegisterFile_WithContentItem_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusCreated}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When RegisterFile is called with content_item data", func() {
			isPublishable := false
			metadata := files.StoredRegisteredMetaData{
				Path:          "datasets/cpih/2024/data.csv",
				IsPublishable: isPublishable,
				Title:         "CPIH Dataset",
				SizeInBytes:   54321,
				Type:          "text/csv",
				Licence:       "OGL v3",
				LicenceURL:    "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
				ContentItem: &files.StoredContentItem{
					DatasetID: "dataset-123",
					Edition:   "2024",
					Version:   "1",
				},
			}

			err := client.RegisterFile(context.Background(), metadata, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPost)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
				So(actualCall.Req.Header.Get("Content-Type"), ShouldEqual, "application/json")

				bodyBytes, err := io.ReadAll(actualCall.Req.Body)
				So(err, ShouldBeNil)

				var actualMetadata files.StoredRegisteredMetaData
				err = json.Unmarshal(bodyBytes, &actualMetadata)
				So(err, ShouldBeNil)
				So(actualMetadata, ShouldResemble, metadata)
				So(actualMetadata.ContentItem, ShouldNotBeNil)
				So(actualMetadata.ContentItem.DatasetID, ShouldEqual, "dataset-123")
				So(actualMetadata.ContentItem.Edition, ShouldEqual, "2024")
				So(actualMetadata.ContentItem.Version, ShouldEqual, "1")
			})
		})
	})
}

func TestRegisterFile_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that fails when Do() is called", t, func() {
		mockClienter := newMockClienter(nil, errExpectedDoFailure)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When RegisterFile is called", func() {
			isPublishable := true
			metadata := files.StoredRegisteredMetaData{
				Path:          "path/to/file.txt",
				IsPublishable: isPublishable,
				Title:         "Test File",
				SizeInBytes:   12345,
				Type:          "text/plain",
				Licence:       "OGL v3",
				LicenceURL:    "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
			}

			err := client.RegisterFile(context.Background(), metadata, testHeaders)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, errExpectedDoFailure)
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})

	Convey("Given a files-api client that returns an unexpected status code", t, func() {
		body := `{"errors":[{"errorCode":"ValidationError","description":"Path required"}]}`
		mockClienter := newMockClienter(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When RegisterFile is called", func() {
			isPublishable := true
			metadata := files.StoredRegisteredMetaData{
				Path:          "path/to/file.txt",
				IsPublishable: isPublishable,
				Title:         "Test File",
				SizeInBytes:   12345,
				Type:          "text/plain",
				Licence:       "OGL v3",
				LicenceURL:    "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
			}

			err := client.RegisterFile(context.Background(), metadata, testHeaders)

			Convey("Then an APIError is returned with the expected status code and errors", func() {
				expectedError := &APIError{
					StatusCode: http.StatusBadRequest,
					Errors: &api.JSONErrors{
						Error: []api.JSONError{
							{
								Code:        "ValidationError",
								Description: "Path required",
							},
						},
					},
				}
				So(err, ShouldResemble, expectedError)
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestRegisterFile_ErrorResponse(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that returns a non-JSON error response", t, func() {
		mockClienter := newMockClienter(&http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When RegisterFile is called", func() {
			isPublishable := true
			metadata := files.StoredRegisteredMetaData{
				Path:          "path/to/file.txt",
				IsPublishable: isPublishable,
				Title:         "Test File",
				SizeInBytes:   12345,
				Type:          "text/plain",
				Licence:       "OGL v3",
				LicenceURL:    "http://example.com/licence",
			}

			err := client.RegisterFile(context.Background(), metadata, testHeaders)

			Convey("Then an APIError is returned with the status code but no parsed errors", func() {
				apiErr, ok := err.(*APIError)
				So(ok, ShouldBeTrue)
				So(apiErr.StatusCode, ShouldEqual, http.StatusUnauthorized)
				So(apiErr.Errors, ShouldBeNil)
			})
		})
	})
}
