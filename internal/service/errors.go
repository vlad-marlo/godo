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
	ErrForbidden = fielderr.New(
		"user has no permission",
		map[string]string{
			"error": "you have no permission to do",
		},
		fielderr.CodeForbidden)
	ErrGroupAlreadyExists = fielderr.New(
		"group already exists",
		map[string]string{
			"name": "group with provided name already exists",
		},
		fielderr.CodeConflict,
	)
	ErrConflict           = fielderr.New("conflict", nil, fielderr.CodeConflict)
	ErrBadData            = fielderr.New("bad data", nil, fielderr.CodeBadRequest)
	ErrBadAuthCredentials = fielderr.New("bad auth credentials", nil, fielderr.CodeUnauthorized)
	ErrNotFound           = fielderr.New("not found", nil, fielderr.CodeNotFound)
	ErrBadInvite          = fielderr.New(
		"invite does not exists",
		map[string]string{
			"invite": "is not available",
		},
		fielderr.CodeNotFound,
	)
	ErrAlreadyInGroup = fielderr.New(
		"user already in group",
		map[string]string{
			"invite": "you are already in group",
		},
		fielderr.CodeConflict,
	)
	ErrUserNotFound = fielderr.New("user not found", map[string]string{
		"user": "not found",
	}, fielderr.CodeNotFound)
	ErrBadInviteLimit = fielderr.New("bad limit", map[string]string{
		"limit": "limit must be not null positive integer number",
	}, fielderr.CodeBadRequest)
)
