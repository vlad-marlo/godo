package pgx

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
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
	if _, err := repo.pool.Exec(ctx, `INSERT INTO invites(id, group_id, role_id, use_count)
VALUES ($1,
        $2,
        (SELECT x.id FROM roles x WHERE x.tasks = $4 AND x.members = $5 AND x.reviews = $6 AND x.comments = $7 LIMIT 1),
        $3);`, invite, group, uses, r.Tasks, r.Members, r.Reviews, r.Comments); err != nil {

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
