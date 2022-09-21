package store

import "errors"

var (
	ErrDuplicateFile              = errors.New("duplicate file path")
	ErrFileNotRegistered          = errors.New("file not registered")
	ErrFileNotInCreatedState      = errors.New("file state is not in state created")
	ErrFileNotInUploadedState     = errors.New("file state is not in state uploaded")
	ErrFileNotInPublishedState    = errors.New("file state is not in state published")
	ErrFileIsNotPublishable       = errors.New("file is not set as publishable")
	ErrNoFilesInCollection        = errors.New("no files found in collection")
	ErrCollectionIDAlreadySet     = errors.New("collection ID already set")
	ErrCollectionIDNotSet         = errors.New("collection ID not set")
	ErrCollectionAlreadyPublished = errors.New("collection with the given id is already published")
)
