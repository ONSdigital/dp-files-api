package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/store"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errExpectedDoFailure   = errors.New("intentional Do error")
	errExpectedReadFailure = errors.New("intentional read error")
	brokenReader           = io.NopCloser(iotest.ErrReader(errExpectedReadFailure))

	exampleCollectionID = "collection1"
	exampleBundleID     = "bundle1"
	exampleTime         = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	exampleEtag         = "example-etag"

	exampleStoredRegisteredMetaData = &files.StoredRegisteredMetaData{
		Path:              "path/to/file.txt",
		IsPublishable:     true,
		CollectionID:      &exampleCollectionID,
		BundleID:          &exampleBundleID,
		Title:             "Title of Metadata",
		SizeInBytes:       123,
		Type:              "text/csv",
		Licence:           "Example Licence",
		LicenceURL:        "http://example.com/licence",
		CreatedAt:         exampleTime,
		LastModified:      exampleTime.Add(time.Hour),
		UploadCompletedAt: &exampleTime,
		PublishedAt:       nil,
		MovedAt:           nil,
		State:             store.StateUploaded,
		Etag:              exampleEtag,
	}
)

func TestUnmarshalJSONErrors(t *testing.T) {
	t.Parallel()

	Convey("Given a valid JSONErrors body", t, func() {
		body := `{"errors":[{"errorCode":"FileNotRegistered","description":"file not registered"}]}`
		rc := io.NopCloser(strings.NewReader(body))

		Convey("When unmarshalJSONErrors is called", func() {
			jsonErrors, err := unmarshalJSONErrors(context.Background(), rc)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected JSONErrors is returned", func() {
				expected := &api.JSONErrors{
					Error: []api.JSONError{
						{
							Code:        "FileNotRegistered",
							Description: "file not registered",
						},
					},
				}
				So(jsonErrors, ShouldResemble, expected)
			})
		})
	})

	Convey("Given an invalid JSONErrors body", t, func() {
		body := `invalid json`
		rc := io.NopCloser(strings.NewReader(body))

		Convey("When unmarshalJSONErrors is called", func() {
			jsonErrors, err := unmarshalJSONErrors(context.Background(), rc)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And no JSONErrors is returned", func() {
				So(jsonErrors, ShouldBeNil)
			})
		})
	})

	Convey("Given a nil body", t, func() {
		Convey("When unmarshalJSONErrors is called", func() {
			jsonErrors, err := unmarshalJSONErrors(context.Background(), nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And no JSONErrors is returned", func() {
				So(jsonErrors, ShouldBeNil)
			})
		})
	})

	Convey("Given a reader that returns an error on read", t, func() {
		Convey("When unmarshalJSONErrors is called", func() {
			jsonErrors, err := unmarshalJSONErrors(context.Background(), brokenReader)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, errExpectedReadFailure)
			})

			Convey("And no JSONErrors is returned", func() {
				So(jsonErrors, ShouldBeNil)
			})
		})
	})

}

func TestUnmarshalStoredRegisteredMetaData(t *testing.T) {
	t.Parallel()

	Convey("Given a valid StoredRegisteredMetaData body", t, func() {
		body, err := json.Marshal(exampleStoredRegisteredMetaData)
		So(err, ShouldBeNil)
		rc := io.NopCloser(strings.NewReader(string(body)))

		Convey("When unmarshalStoredRegisteredMetaData is called", func() {
			metadata, err := unmarshalStoredRegisteredMetaData(rc)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected StoredRegisteredMetaData is returned", func() {
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
		})
	})

	Convey("Given an invalid StoredRegisteredMetaData body", t, func() {
		body := `invalid json`
		rc := io.NopCloser(strings.NewReader(body))

		Convey("When unmarshalStoredRegisteredMetaData is called", func() {
			metadata, err := unmarshalStoredRegisteredMetaData(rc)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And no StoredRegisteredMetaData is returned", func() {
				So(metadata, ShouldBeNil)
			})
		})
	})

	Convey("Given a nil body", t, func() {
		Convey("When unmarshalStoredRegisteredMetaData is called", func() {
			metadata, err := unmarshalStoredRegisteredMetaData(nil)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, ErrMissingResponseBody)
			})

			Convey("And no StoredRegisteredMetaData is returned", func() {
				So(metadata, ShouldBeNil)
			})
		})
	})

	Convey("Given a reader that returns an error on read", t, func() {
		Convey("When unmarshalStoredRegisteredMetaData is called", func() {
			metadata, err := unmarshalStoredRegisteredMetaData(brokenReader)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, errExpectedReadFailure)
			})

			Convey("And no StoredRegisteredMetaData is returned", func() {
				So(metadata, ShouldBeNil)
			})
		})
	})
}
