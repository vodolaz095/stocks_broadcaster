package service

import (
	"errors"
)

var (
	DuplicateSubscriberError = errors.New("duplicate subscriber")
	NoWritersError           = errors.New("no writers defined")
	NoReadersError           = errors.New("no readers defined")
)
