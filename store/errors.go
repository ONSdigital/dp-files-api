package store

import "errors"

var (
	ErrDuplicateFile                   = errors.New("duplicate file path")
	ErrFileNotRegistered               = errors.New("file not registered")
	ErrFileNotInCreatedState           = errors.New("file state is not in state created")
	ErrFileNotInUploadedState          = errors.New("file state is not in state uploaded")
	ErrFileStateMismatch               = errors.New("file state mismatch")
	ErrFileIsNotPublishable            = errors.New("file is not set as publishable")
	ErrNoFilesInCollection             = errors.New("no files found in collection")
	ErrCollectionIDAlreadySet          = errors.New("collection ID already set")
	ErrCollectionIDNotSet              = errors.New("collection ID not set")
	ErrCollectionAlreadyPublished      = errors.New("collection with the given id is already published")
	ErrBundleAlreadyPublished          = errors.New("bundle with the given id is already published")
	ErrCollectionMetadataNotRegistered = errors.New("collection metadata not registered")
	ErrBundleMetadataNotRegistered     = errors.New("bundle metadata not registered")
	ErrEtagMismatchWhilePublishing     = errors.New("etag mismatch")
	ErrBundleIDAlreadySet              = errors.New("bundle ID already set")
)
