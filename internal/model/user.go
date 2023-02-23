package model

import (
	"github.com/google/uuid"
)

type (
	// User ...
	User struct {
		// ID is user uuid.
		ID uuid.UUID `json:"id"`
		// Email is string field of user's email addr.
		Email string `json:"email"`
		// Pass is encrypted user password.
		Pass string `json:"-"`
	}
	// RegisterUserRequest ...
	RegisterUserRequest struct {
		// Email is user email
		Email string `json:"email"`
		// Password is password string
		Password string `json:"password"`
	}

	// LoginUserRequest ...
	LoginUserRequest struct {
		// Email is user email
		Email string `json:"email"`
		// Password is password of user.
		Password string `json:"password"`
	}

	// CreateUserResponse ...
	CreateUserResponse struct {
		ID    uuid.UUID `json:"id"`
		Email string    `json:"email"`
	}

	// CreateJWTResponse is request object which will return to user on token create.
	CreateJWTResponse struct {
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)
