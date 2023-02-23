//go:generate mockgen --source=pgx.go --destination=mocks/mock.go --package=mocks

package pgx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/store"
)

var _ store.Store = &Store{}

// Store is implementation of storage Interface.
type Store struct {
	p     *pgxpool.Pool
	l     *zap.Logger
	user  *UserRepository
	group *GroupRepository
}

type Client interface {
	// P returns prepared pgx pool connection to database.
	P() *pgxpool.Pool
	// L returns prepared storage zap log.
	L() *zap.Logger
}

// New ...
func New(client Client, user *UserRepository, group *GroupRepository) *Store {
	return &Store{
		p:     client.P(),
		l:     client.L(),
		user:  user,
		group: group,
	}
}

// User returns user repository.
func (store *Store) User() store.UserRepository {
	return store.user
}

// Group return group repository.
func (store *Store) Group() store.GroupRepository {
	return store.group
}

// Ping checks connection to database.
func (store *Store) Ping(ctx context.Context) error {
	return store.p.Ping(ctx)
}

func (store *Store) Close() {
	if store != nil {
		store.p.Close()
	}
}
