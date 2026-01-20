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

func TestPatchFile_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When PatchFile is called with a leading slash", func() {
			// in reality all these fields would not be set together, they are all set here for test completeness
			patchReq := FilePatchRequest{
				StateMetadata: api.StateMetadata{
					State:        stringToPointer("PUBLISHED"),
					CollectionID: stringToPointer(exampleCollectionID),
					BundleID:     stringToPointer(exampleBundleID),
				},
				ETag: exampleEtag,
			}

			err := client.patchFile(context.Background(), "/path/to/file.txt", patchReq, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPatch)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)

				bodyBytes, err := io.ReadAll(actualCall.Req.Body)
				So(err, ShouldBeNil)

				var actualPatchReq FilePatchRequest
				err = json.Unmarshal(bodyBytes, &actualPatchReq)
				So(err, ShouldBeNil)
				So(actualPatchReq, ShouldResemble, patchReq)
			})
		})

		Convey("When PatchFile is called without a leading slash", func() {
			// in reality all these fields would not be set together, they are all set here for test completeness
			patchReq := FilePatchRequest{
				StateMetadata: api.StateMetadata{
					State:        stringToPointer("PUBLISHED"),
					CollectionID: stringToPointer(exampleCollectionID),
					BundleID:     stringToPointer(exampleBundleID),
				},
				ETag: exampleEtag,
			}

			err := client.patchFile(context.Background(), "path/to/file.txt", patchReq, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPatch)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)

				bodyBytes, err := io.ReadAll(actualCall.Req.Body)
				So(err, ShouldBeNil)

				var actualPatchReq FilePatchRequest
				err = json.Unmarshal(bodyBytes, &actualPatchReq)
				So(err, ShouldBeNil)
				So(actualPatchReq, ShouldResemble, patchReq)
			})
		})
	})
}

func TestPatchFile_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client that fails when Do() is called", t, func() {
		mockClienter := newMockClienter(nil, errExpectedDoFailure)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When PatchFile is called", func() {
			patchReq := FilePatchRequest{}

			err := client.patchFile(context.Background(), "/path/to/file.txt", patchReq, testHeaders)

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

		Convey("When PatchFile is called", func() {
			patchReq := FilePatchRequest{}

			err := client.patchFile(context.Background(), "/path/to/file.txt", patchReq, testHeaders)

			Convey("Then an APIError is returned with the expected status code and errors", func() {
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
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

// Only testing that MarkFilePublished calls patchFile with the correct parameters
// all other logic is within patchFile which is tested separately
func TestMarkFilePublished(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When MarkFilePublished is called", func() {
			err := client.MarkFilePublished(context.Background(), "/path/to/file.txt", testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPatch)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)

				bodyBytes, err := io.ReadAll(actualCall.Req.Body)
				So(err, ShouldBeNil)

				var actualPatchReq FilePatchRequest
				err = json.Unmarshal(bodyBytes, &actualPatchReq)
				So(err, ShouldBeNil)
				expectedPatchReq := FilePatchRequest{
					StateMetadata: api.StateMetadata{
						State: stringToPointer("PUBLISHED"),
					},
				}
				So(actualPatchReq, ShouldResemble, expectedPatchReq)
			})
		})
	})
}

// Only testing that MarkFileUploaded calls patchFile with the correct parameters
// all other logic is within patchFile which is tested separately
func TestMarkFileUploaded(t *testing.T) {
	t.Parallel()

	Convey("Given a files-api client", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK}, nil)
		client := newMockFilesAPIClient(mockClienter)

		Convey("When MarkFileUploaded is called", func() {
			err := client.MarkFileUploaded(context.Background(), "/path/to/file.txt", exampleEtag, testHeaders)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPatch)
				So(actualCall.Req.URL.String(), ShouldEqual, filesAPIURL+"/files/path/to/file.txt")
				So(actualCall.Req.Header.Get("Authorization"), ShouldEqual, "Bearer "+testAuthToken)

				bodyBytes, err := io.ReadAll(actualCall.Req.Body)
				So(err, ShouldBeNil)

				var actualPatchReq FilePatchRequest
				err = json.Unmarshal(bodyBytes, &actualPatchReq)
				So(err, ShouldBeNil)
				expectedPatchReq := FilePatchRequest{
					StateMetadata: api.StateMetadata{
						State: stringToPointer("UPLOADED"),
					},
					ETag: exampleEtag,
				}
				So(actualPatchReq, ShouldResemble, expectedPatchReq)
			})
		})
	})
}
