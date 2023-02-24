package pgx

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
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
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		repo.log.Error("failed to begin transaction: check pgx driver", TraceError(err)...)
		return store.ErrUnknown
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			repo.log.Error("failed to rollback transaction: check pgx driver", TraceError(err)...)
		}
	}()
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO groups(id, name, description, owner) VALUES ($1, $2, $3, $4) RETURNING created_at;`,
		group.ID,
		group.Name,
		group.Description,
		group.Owner,
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

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO roles(members, tasks, reviews, "comments")  VALUES (0, 0, 0, 0) ON CONFLICT DO NOTHING;`,
	); err != nil {

		repo.log.Error("unknown error while creating role", TraceError(err)...)
		return store.ErrUnknown
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO user_in_group(user_id, group_id, role_id, is_admin)
VALUES ($1, $2, (SELECT x.id
                 FROM roles x
                 WHERE x.comments = 0
                   AND x.reviews = 0
                   AND x.members = 0
                   AND x.tasks = 0
                 LIMIT 1), $3);`,
		group.Owner,
		group.ID,
		true,
	); err != nil {
		repo.log.Error("unknown error while creating record about user in group", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	if err := tx.Commit(ctx); err != nil {
		repo.log.Error("failed to commit transaction: check pgx driver", TraceError(err)...)
		return store.ErrUnknown
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
		&g.Owner,
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

func (repo *GroupRepository) GetUsers(ctx context.Context, id string) ([]uuid.UUID, error) {
	rows, err := repo.pool.Query(ctx, `SELECT user_id FROM user_in_group WHERE group_id = $1;`, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			repo.log.Error("unknown error while scanning rows", TraceError(err)...)
			return nil, store.ErrUnknown
		}
		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		repo.log.Error("error occurred while reading query", TraceError(err)...)
		return nil, fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	return ids, nil
}
