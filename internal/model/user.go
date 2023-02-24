package model

import (
	"github.com/google/uuid"
)

type (
	// User ...
	User struct {
		// ID is user uuid.
		ID uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000000"`
		// Email is string field of user's email addr.
		Email string `json:"email" example:"user@example.com"`
		// Pass is encrypted user password.
		Pass string `json:"-"`
	}

	UserInGroup struct {
		UserID  uuid.UUID
		IsAdmin bool
		Members int
		Tasks   int
		Reviews int
		Comment int
	}
	// RegisterUserRequest ...
	RegisterUserRequest struct {
		// Email is user email
		Email string `json:"email" example:"user@example.com"`
		// Password is password string
		Password string `json:"password" example:"strong_password"`
	}

	// CreateTokenRequest ...
	CreateTokenRequest struct {
		// Email is user email
		Email string `json:"email" example:"user@example.com"`
		// Password is password of user.
		Password  string `json:"password" example:"strong_password"`
		TokenType string `json:"token-type" example:"bearer"`
	}

	// CreateUserResponse ...
	CreateUserResponse struct {
		ID    uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000000"`
		Email string    `json:"email" example:"user@example.com"`
	}

	// CreateTokenResponse is request object which will return to user on token create.
	CreateTokenResponse struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
)
