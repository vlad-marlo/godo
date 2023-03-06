package store

import (
	"errors"
)

var (
	// ErrUserAlreadyExists is unique violation error.
	ErrUserAlreadyExists   = errors.New("user already exists with provided username")
	ErrBadData             = errors.New("bad data")
	ErrNotFound            = errors.New("not found")
	ErrUnknown             = errors.New("unknown error")
	ErrTokenAlreadyExists  = errors.New("token already exists")
	ErrTaskAlreadyExists   = errors.New("task already exists")
	ErrInviteIsAlreadyUsed = errors.New("invite was already used")
	ErrUniqueViolation     = errors.New("unique violation")
	ErrFKViolation         = errors.New("foreign key violation")
)
