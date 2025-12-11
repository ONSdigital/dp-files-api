package api

import (
	"testing"

	"github.com/go-playground/validator"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAWSUploadKeyValidator(t *testing.T) {
	t.Parallel()

	type fields struct {
		Path string `validate:"aws-upload-key"`
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"starts with lowercase letter", "file.csv", true},
		{"starts with uppercase letter", "File.csv", true},
		{"starts with number", "1file.csv", true},
		{"uuid format", "123e4567-e89b-12d3-a456-426614174000/file.csv", true},
		{"empty string", "", false},
		{"starts with slash", "/file.csv", false},
		{"starts with hyphen", "-file.csv", false},
		{"starts with underscore", "_file.csv", false},
	}

	Convey("Given the following AWS upload key validation cases", t, func() {
		validate := validator.New()
		err := validate.RegisterValidation("aws-upload-key", awsUploadKeyValidator)
		So(err, ShouldBeNil)

		for _, tc := range tests {
			Convey(tc.name, func() {
				field := fields{Path: tc.path}
				err := validate.Struct(field)
				if tc.expected {
					So(err, ShouldBeNil)
				} else {
					So(err, ShouldNotBeNil)
				}
			})
		}
	})
}
