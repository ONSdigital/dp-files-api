package files

import "errors"

var (
	ErrDuplicateFile           = errors.New("duplicate file path")
	ErrFileNotRegistered       = errors.New("file not registered")
	ErrFileNotInCreatedState   = errors.New("file state is not in state created")
	ErrFileNotInUploadedState  = errors.New("file state is not in state uploaded")
	ErrFileNotInPublishedState = errors.New("file state is not in state published")
	ErrNoFilesInCollection     = errors.New("no files found in collection")
	ErrCollectionIDAlreadySet  = errors.New("collection ID already set")
	ErrCollectionIDNotSet      = errors.New("collection ID not set")
)
