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
	// GetByEmail return user with provided email.
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	// Exists return existence of user with provided id.
	Exists(ctx context.Context, id string) (ok bool)
	// Get return user with provided id.
	Get(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// GroupRepository give user access to group storage - Create, Update, Delete, check existence of groups.
type GroupRepository interface {
	// Create creates record about group if it does not exist.
	Create(ctx context.Context, group *model.Group) error
	// Exists returns existence of group with provided id.
	//Exists(ctx context.Context, id string) (ok bool)

	// AddTask create relation between task and group
	AddTask(ctx context.Context, task, group string) error
	// UserExists checks existence of user in group members.
	UserExists(ctx context.Context, group, user string) (ok bool)
	// GetByUser ...
	GetByUser(ctx context.Context, user uuid.UUID) ([]*model.Group, error)
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
	// Create creates invite with provided data.
	Create(ctx context.Context, invite uuid.UUID, r *model.Role, group uuid.UUID, uses int) error
	// Exists checks existence valid invite with provided data.
	Exists(ctx context.Context, invite, group uuid.UUID) bool
	// Use decrements left uses of invite and adds user to group in tx.
	Use(ctx context.Context, invite uuid.UUID, user uuid.UUID) error
}

// TaskRepository ...
type TaskRepository interface {
	// AllByGroupAndUser return all tasks that are related to group.
	AllByGroupAndUser(ctx context.Context, group uuid.UUID, user uuid.UUID) ([]*model.Task, error)
	// AllByUser ...
	AllByUser(ctx context.Context, user uuid.UUID) ([]*model.Task, error)
	// GetByUserAndID return task that has id task and is related to user.
	GetByUserAndID(ctx context.Context, user, task uuid.UUID) (*model.Task, error)
	// Create creates record about task.
	Create(ctx context.Context, task *model.Task) error
	// AddToUser ...
	AddToUser(ctx context.Context, from, task, to uuid.UUID) error
	// AddToGroup ...
	AddToGroup(ctx context.Context, task, group uuid.UUID) error
}

// Store is composite object that does not include any storage function.
//
// Store is only accessor to different repositories.
type Store interface {
	// User is UserRepository accessor.
	User() UserRepository
	// Group is GroupRepository accessor.
	Group() GroupRepository
	// Token is TokenRepository accessor.
	Token() TokenRepository
	// Task is TaskRepository accessor.
	Task() TaskRepository
	// Invite is InviteRepository accessor.
	Invite() InviteRepository
	// Ping checks is Store working correctly.
	Ping(ctx context.Context) error
}
