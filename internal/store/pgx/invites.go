package pgx

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

var _ store.InviteRepository = (*InviteRepository)(nil)

type InviteRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewInviteRepository create new invite repository that encapsulates logic to store invites.
func NewInviteRepository(cli Client) *InviteRepository {
	return &InviteRepository{
		pool: cli.P(),
		log:  cli.L(),
	}
}

// Create stores invite with provided data.
func (repo *InviteRepository) Create(ctx context.Context, invite uuid.UUID, r *model.Role, group uuid.UUID, uses int) error {
	if r == nil {
		return store.ErrNilReference
	}

	if _, err := repo.pool.Exec(
		ctx,
		`INSERT INTO roles(members, tasks, reviews, "comments")
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING`,
		r.Members,
		r.Tasks,
		r.Reviews,
		r.Comments,
	); err != nil {
		repo.log.Log(_unknownLevel, "create role", TraceError(err)...)
		return Unknown(err)
	}
	if err := repo.pool.QueryRow(ctx, `SELECT id FROM roles WHERE members = $1 AND tasks = $2 AND reviews = $3 AND "comments" = $4;`, r.Members, r.Tasks, r.Reviews, r.Comments).Scan(&r.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		repo.log.Log(_unknownLevel, "get role id", TraceError(err)...)
		return Unknown(err)
	}

	if _, err := repo.pool.Exec(
		ctx,
		`INSERT INTO invites(id, group_id, role_id, use_count)
VALUES ($1, $2, $3, $4);`,
		invite,
		group,
		r.ID,
		uses,
	); err != nil {

		if pgErr, ok := err.(*pgconn.PgError); ok {

			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrUniqueViolation
			case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
				return store.ErrFKViolation
			case pgerrcode.CheckViolation:
				return store.ErrBadData
			}
		}
		repo.log.Warn("unknown error while creating invite link", TraceError(err)...)
		return Unknown(err)
	}

	return nil
}

// Exists checks existence of invite to group with data.
func (repo *InviteRepository) Exists(ctx context.Context, invite, group uuid.UUID) (ok bool) {
	if err := repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM invites WHERE id = $1 AND group_id = $2 AND use_count > 0);`,
		invite,
		group,
	).Scan(&ok); err != nil {
		repo.log.Warn("unknown error while getting", TraceError(err)...)
	}
	return ok
}

// Use add user to group if invite is right.
func (repo *InviteRepository) Use(ctx context.Context, invite uuid.UUID, user uuid.UUID) error {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		repo.log.Error("unexpected error received while starting new transaction: check drivers", TraceError(err)...)
		return Unknown(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var group uuid.UUID
	var role int

	if err = tx.QueryRow(
		ctx,
		`UPDATE invites SET use_count = use_count - 1 WHERE id = $1 RETURNING group_id, role_id;`,
		invite,
	).Scan(
		&group,
		&role,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.CheckViolation {
				return store.ErrInviteIsAlreadyUsed
			}
		}
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO user_in_group(user_id, group_id, role_id) VALUES ($1, $2, $3);`,
		user,
		group,
		role,
	); err != nil {

		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {

			case pgerrcode.ForeignKeyViolation:
				return store.ErrBadData

			case pgerrcode.UniqueViolation:
				return store.ErrInviteIsAlreadyUsed
			}

			repo.log.Error("add user to group", TraceError(err)...)
		}

		return Unknown(err)
	}

	if err = tx.Commit(ctx); err != nil {
		repo.log.Error("unexpected error while doing commit transaction: check pgx driver", TraceError(err)...)
		return Unknown(err)
	}

	return nil
}
