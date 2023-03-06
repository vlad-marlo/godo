package service

import (
	"context"
	"github.com/google/uuid"
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
	GetUserFromToken(ctx context.Context, t string) (uuid.UUID, error)
	// CreateGroup create new group.
	CreateGroup(ctx context.Context, user uuid.UUID, name, description string) (*model.CreateGroupResponse, error)
	// CreateInvite creates invite link on which user will insert into group.
	CreateInvite(ctx context.Context, user uuid.UUID, group uuid.UUID, role *model.Role, limit int) (*model.CreateInviteResponse, error)
	// UseInvite applies use to group if invite data is ok.
	UseInvite(ctx context.Context, user uuid.UUID, group uuid.UUID, invite uuid.UUID) error
	// GetMe ...
	GetMe(ctx context.Context, user uuid.UUID) (*model.GetMeResponse, error)
	// GetUserTasks return all tasks, related to user.
	GetUserTasks(ctx context.Context, user uuid.UUID) (*model.GetTasksResponse, error)
	// GetTask return task by id if user is related to it.
	GetTask(ctx context.Context, user, task uuid.UUID) (*model.Task, error)
	// CreateTask ...
	CreateTask(ctx context.Context, user uuid.UUID, task model.TaskCreateRequest) (*model.Task, error)
}
