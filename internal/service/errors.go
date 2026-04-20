package service

import (
	"errors"
)

var (
	ErrDuplicateSubscriber = errors.New("duplicate subscriber")
	ErrNoWriters           = errors.New("no writers defined")
	ErrNoReaders           = errors.New("no readers defined")
)
