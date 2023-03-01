package service

import (
	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
)

var (
	ErrTokenNotValid = fielderr.New(
		"token is not valid",
		map[string]string{
			"token": "is not valid token",
		},
		fielderr.CodeUnauthorized,
	)
	ErrBadTokenType = fielderr.New(
		"bad token type",
		map[string]string{
			"token-type": "token type must be auth-token, bearer or jwt",
		},
		fielderr.CodeBadRequest,
	)
	ErrEmailAlreadyInUse = fielderr.New(
		"email already in use",
		map[string]string{
			"email": "already in use",
		},
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
	ErrBadUser = fielderr.New(
		"bad user id",
		map[string]string{
			"user": "check auth credentials",
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
