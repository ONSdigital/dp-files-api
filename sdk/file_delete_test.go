package sdk

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteFile_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusNoContent}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When DeleteFile is called with a leading slash", func() {
			err := client.DeleteFile(context.Background(), "/path/to/file.txt", testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodDelete)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})

		Convey("When DeleteFile is called without a leading slash", func() {
			err := client.DeleteFile(context.Background(), "path/to/file.txt", testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodDelete)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)
			})
		})
	})
}

func TestDeleteFile_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that fails when Do() is called", t, func() {
		mockClienter := newMockClienter(nil, errExpectedDoFailure)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When DeleteFile is called", func() {
			err := client.DeleteFile(context.Background(), "/path/to/file.txt", testHeaders)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, errExpectedDoFailure)
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

		Convey("When DeleteFile is called", func() {
			err := client.DeleteFile(context.Background(), "/path/to/file.txt", testHeaders)

			Convey("Then the expected APIError is returned", func() {
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

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})
}
