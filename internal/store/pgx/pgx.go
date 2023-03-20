//go:generate mockgen --source=pgx.go --destination=mocks/mock.go --package=mocks

package pgx

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/vlad-marlo/godo/internal/store"
)

var _ store.Store = (*Store)(nil)

const _unknownLevel = zapcore.WarnLevel

// Store is implementation of storage Interface.
type Store struct {
	pool   *pgxpool.Pool
	log    *zap.Logger
	user   *UserRepository
	group  *GroupRepository
	token  *TokenRepository
	task   *TaskRepository
	invite *InviteRepository
	role   *RoleRepository
}

type Client interface {
	// P returns prepared pgx pool connection to database.
	P() *pgxpool.Pool
	// L returns prepared storage zap log.
	L() *zap.Logger
}

// New ...
func New(
	client Client,
	user *UserRepository,
	group *GroupRepository,
	token *TokenRepository,
	task *TaskRepository,
	invite *InviteRepository,
	role *RoleRepository,
) *Store {
	return &Store{
		pool:   client.P(),
		log:    client.L(),
		user:   user,
		group:  group,
		token:  token,
		task:   task,
		invite: invite,
		role:   role,
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

// Token return token repository.
func (store *Store) Token() store.TokenRepository {
	return store.token
}

// Task return task repository
func (store *Store) Task() store.TaskRepository {
	return store.task
}

// Invite return invite repository.
func (store *Store) Invite() store.InviteRepository {
	return store.invite
}

// Ping checks connection to database.
func (store *Store) Ping(ctx context.Context) error {
	return store.pool.Ping(ctx)
}

func (store *Store) Role() store.RoleRepository {
	return store.role
}

// Close is helper function to close connection.
func (store *Store) Close() {
	if store != nil {
		store.pool.Close()
	}
}
