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
	pool  *pgxpool.Pool
	log   *zap.Logger
	user  *UserRepository
	group *GroupRepository
	token *TokenRepository
	task  *TaskRepository
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
) *Store {
	return &Store{
		pool:  client.P(),
		log:   client.L(),
		user:  user,
		group: group,
		token: token,
		task:  task,
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

func (store *Store) Token() store.TokenRepository {
	return store.token
}

func (store *Store) Task() store.TaskRepository {
	return store.task
}

// Ping checks connection to database.
func (store *Store) Ping(ctx context.Context) error {
	return store.pool.Ping(ctx)
}

func (store *Store) Close() {
	if store != nil {
		store.pool.Close()
	}
}
