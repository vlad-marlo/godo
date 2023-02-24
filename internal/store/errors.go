package store

import (
	"errors"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists with provided username")
	ErrGroupAlreadyExists = errors.New("group already exists with provided data")
	ErrBadData            = errors.New("bad data")
	ErrNotFound           = errors.New("not found")
	ErrUnknown            = errors.New("unknown error")
	ErrBadForeignKey      = errors.New("bad foreign key")
)
