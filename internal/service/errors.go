package service

import (
	"errors"
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
)

var (
	ErrTokenNotValid     = errors.New("token is not valid")
	ErrLoginAlreadyInUse = fielderr.New(
		"login already in use",
		map[string]string{"login": "already in use"},
		fielderr.CodeConflict,
	)
	ErrInternal = fielderr.New(
		"internal server error",
		"internal server error",
		fielderr.CodeInternal,
	)
)
