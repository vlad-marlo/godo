//go:generate mockgen --source=interface.go --destination=mocks/store.go --package=mocks

package store

import (
	"context"
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
	// AddUser adding user to group.
	AddUser(ctx context.Context, invite string, user string) error
	// AddTask create relation between task and group
	AddTask(ctx context.Context, task, group string) error
}

// TokenRepository is accessor to storing tokens.
type TokenRepository interface {
	// Create creates unique token.
	Create(ctx context.Context, token *model.Token) error
	// Get return token with provided token string.
	Get(ctx context.Context, token string) (*model.Token, error)
}

// TaskRepository ...
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
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
	// Ping checks is Store working correctly.
	Ping(ctx context.Context) error
}
