package pgx

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

type RoleRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewRoleRepository(cli Client) *RoleRepository {
	return &RoleRepository{
		pool: cli.P(),
		log:  cli.L(),
	}
}

func (repo *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	if role == nil {
		return store.ErrNilReference
	}
	if err := repo.pool.QueryRow(
		ctx,
		`INSERT INTO roles(members, tasks, reviews, comments) VALUES ($1, $2, $3, $4) RETURNING id;`,
		role.Members,
		role.Tasks,
		role.Reviews,
		role.Comments,
	).Scan(&role.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrUniqueViolation
			default:

			}
		}
		zap.L().Log(_unknownLevel, "store: role: create", traceError(err)...)

		return unknown(err)
	}
	return nil
}

func (repo *RoleRepository) Get(ctx context.Context, role *model.Role) error {
	if role == nil {
		return store.ErrNilReference
	}
	if err := repo.pool.QueryRow(
		ctx,
		`SELECT id FROM roles WHERE members = $1 AND tasks = $2 AND reviews = $3 AND comments = $4`,
		role.Members,
		role.Tasks,
		role.Reviews,
		role.Comments,
	).Scan(&role.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.Create(ctx, role)
		}
		return unknown(err)
	}
	return nil
}
