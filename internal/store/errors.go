package store

import (
	"errors"
)

var (
	// ErrUserAlreadyExists is unique violation error.
	ErrUserAlreadyExists = errors.New("user already exists with provided username")
	// ErrGroupAlreadyExists is unique violation error.
	ErrGroupAlreadyExists  = errors.New("group already exists with provided data")
	ErrBadData             = errors.New("bad data")
	ErrNotFound            = errors.New("not found")
	ErrUnknown             = errors.New("unknown error")
	ErrTokenIsExpired      = errors.New("token is expired")
	ErrTokenAlreadyExists  = errors.New("token already exists")
	ErrInviteIsAlreadyUsed = errors.New("")
)
