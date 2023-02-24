package service

import (
	"context"
	"github.com/vlad-marlo/godo/internal/model"
)

type Interface interface {
	// Ping checks access to server.
	Ping(ctx context.Context) error
	// CreateToken create new jwt token for refresh and access to server if auth credits are correct.
	CreateToken(ctx context.Context, username, password, token string) (*model.CreateTokenResponse, error)
	// RegisterUser create record about user in storage and prepares response to user.
	RegisterUser(ctx context.Context, email, password string) (*model.User, error)
	// GetUserFromToken is helper function that decodes jwt token from t and check existing of user which id is provided
	// in token claims.
	GetUserFromToken(ctx context.Context, t string) (string, error)
	// CreateGroup create new group.
	CreateGroup(ctx context.Context, user, name, description string) (*model.CreateGroupResponse, error)
}
