package service

import (
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
)

var (
	ErrTokenNotValid = fielderr.New(
		"token is not valid",
		map[string]any{
			"token": "check if it valid",
		},
		fielderr.CodeUnauthorized,
	)
	ErrEmailAlreadyInUse = fielderr.New(
		"email already in use",
		map[string]string{"email": "already in use"},
		fielderr.CodeConflict,
	)
	ErrInternal = fielderr.New(
		"internal server error",
		nil,
		fielderr.CodeInternal,
	)
	ErrPasswordToLong = fielderr.New(
		"password is too long",
		map[string]string{
			"password": "password is too long",
		},
		fielderr.CodeBadRequest,
	)
	ErrPasswordToEasy = fielderr.New(
		"password is to easy",
		map[string]string{
			"password": "password is too easy",
		},
		fielderr.CodeBadRequest,
	)
	ErrEmailNotValid = fielderr.New(
		"bad request",
		map[string]string{
			"email": "pass valid email address",
		},
		fielderr.CodeConflict,
	)
	ErrBadAuthData = fielderr.New(
		"user not found",
		map[string]string{
			"error": "check password or login",
		},
		fielderr.CodeUnauthorized,
	)
	ErrGroupAlreadyExists = fielderr.New(
		"group already exists",
		map[string]string{
			"name": "group with provided name already exists",
		},
		fielderr.CodeConflict,
	)
)
