package pgx

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

var _ store.GroupRepository = &GroupRepository{}

// GroupRepository ...
type GroupRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewGroupRepository return new instance of GroupRepository
func NewGroupRepository(cli Client) *GroupRepository {
	return &GroupRepository{
		pool: cli.P(),
		log:  cli.L(),
	}
}

// Exists ...
func (repo *GroupRepository) Exists(ctx context.Context, id string) (ok bool) {
	_ = repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM groups WHERE id = $1);`,
		id,
	).Scan(&ok)
	return ok
}

// Create ...
func (repo *GroupRepository) Create(ctx context.Context, group *model.Group) error {
	if err := repo.pool.QueryRow(
		ctx,
		`INSERT INTO groups(id, name, description, owner) VALUES ($1, $2, $3, $4) RETURNING created_at;`,
		group.ID,
		group.Name,
		group.Description,
		group.CreatedBy,
	).Scan(&group.CreatedAt); err != nil {

		repo.log.Debug(
			"create user: pgx error",
			TraceError(err)...,
		)

		if pgErr, ok := err.(*pgconn.PgError); ok {

			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrGroupAlreadyExists
			}

			if pgErr.Code == pgerrcode.InvalidForeignKey || pgerrcode.ForeignKeyViolation == pgErr.Code {
				return store.ErrBadData
			}
		}

		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}
	return nil
}

// Get return group with provided id
func (repo *GroupRepository) Get(ctx context.Context, id string) (*model.Group, error) {
	g := new(model.Group)

	if err := repo.pool.QueryRow(
		ctx,
		`SELECT id, name, description, created_at, owner FROM groups WHERE id = $1`,
		id,
	).Scan(
		&g.ID,
		&g.Name,
		&g.Description,
		&g.CreatedAt,
		&g.CreatedBy,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		repo.log.Debug(
			"unknown error while getting group",
			TraceError(err)...,
		)
		return nil, fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}
	return g, nil
}
