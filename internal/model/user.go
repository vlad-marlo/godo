package model

import (
	"github.com/google/uuid"
)

type (
	// User ...
	User struct {
		// ID is user uuid.
		ID uuid.UUID `json:"id"`
		// Name is username.
		Name string `json:"username"`
		// Pass is encrypted user password.
		Pass string `json:"-"`
		// IsAdmin if true then user will have all permissions.
		IsAdmin bool `json:"-"`
	}
	// RegisterUserRequest ...
	RegisterUserRequest struct {
		// Username is username of user.
		Username string `json:"username"`
		// Password is password string
		Password string `json:"password"`
	}

	// LoginUserRequest ...
	LoginUserRequest struct {
		// Username is username of user.
		Username string `json:"username"`
		// Password is password of user.
		Password string `json:"password"`
	}

	// CreateUserResponse ...
	CreateUserResponse struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
	}

	// CreateJWTResponse is request object which will return to user on token create.
	CreateJWTResponse struct {
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)
