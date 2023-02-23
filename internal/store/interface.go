//go:generate mockgen --source=interface.go --destination=mocks/store.go --package=mocks

package store

import (
	"context"

	"github.com/vlad-marlo/godo/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, username string) (*model.User, error)
	Exists(ctx context.Context, user string) bool
}

type GroupRepository interface {
	Create(ctx context.Context, group *model.Group) error
	Exists(ctx context.Context, id string) bool
}

type Store interface {
	User() UserRepository
	Group() GroupRepository
	Ping(ctx context.Context) error
}
