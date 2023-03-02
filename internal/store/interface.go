//go:generate mockgen --source=interface.go --destination=mocks/store.go --package=mocks

package store

import (
	"context"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
)

// UserRepository give user access to storing, getting and changing users.
type UserRepository interface {
	// Create creates new record about user u.
	Create(ctx context.Context, u *model.User) error
	// GetByEmail return user with provided email if exists.
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	// Exists return existence of user with provided id.
	Exists(ctx context.Context, id string) (ok bool)
}

// GroupRepository give user access to group storage - Create, Update, Delete, check existence of groups.
type GroupRepository interface {
	// Create creates record about group if it does not exist.
	Create(ctx context.Context, group *model.Group) error
	// Exists returns existence of group with provided id.
	Exists(ctx context.Context, id string) (ok bool)
	// AddTask create relation between task and group
	AddTask(ctx context.Context, task, group string) error
	// UserExists checks existence of user in group members.
	UserExists(ctx context.Context, group, user string) (ok bool)
	// GetRoleOfMember return role of member in group.
	GetRoleOfMember(ctx context.Context, user, group uuid.UUID) (role *model.Role, err error)
}

// TokenRepository is accessor to storing tokens.
type TokenRepository interface {
	// Create creates unique token.
	Create(ctx context.Context, token *model.Token) error
	// Get return token with provided token string.
	Get(ctx context.Context, token string) (*model.Token, error)
}

type InviteRepository interface {
	Create(ctx context.Context, invite uuid.UUID, r *model.Role, group uuid.UUID, uses int) error
	Exists(ctx context.Context, invite, group uuid.UUID) bool
	Use(ctx context.Context, invite uuid.UUID, user uuid.UUID) error
}

// TaskRepository ...
type TaskRepository interface {
	//Create(ctx context.Context, task *model.Task) error
}

// Store is composite object that does not include any storage function.
// Store is only accessor to different repositories.
type Store interface {
	// User is UserRepository getter.
	User() UserRepository
	// Group is GroupRepository getter.
	Group() GroupRepository
	// Token is TokenRepository getter.
	Token() TokenRepository
	// Task is TaskRepository getter.
	Task() TaskRepository
	// Invite is InviteRepository accessor.
	Invite() InviteRepository
	// Ping checks is Store working correctly.
	Ping(ctx context.Context) error
}
