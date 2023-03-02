package pgx

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
)

var _ store.UserRepository = (*UserRepository)(nil)

// UserRepository ...
type UserRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewUserRepository ...
func NewUserRepository(cli Client) *UserRepository {
	return &UserRepository{cli.P(), cli.L()}
}

// Create creates record about user into db.
//
// repo.RegisterUser(context.Background(), &model.User{ID: uuid.New(), Name: "name", Pass: "password"})
func (repo *UserRepository) Create(
	ctx context.Context,
	u *model.User,
) error {
	if u == nil {
		return store.ErrBadData
	}

	if _, err := repo.pool.Exec(
		ctx,
		`INSERT INTO users (id, email, pass) VALUES ($1, $2, $3);`,
		u.ID,
		u.Email,
		u.Pass,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrUserAlreadyExists
			}
		}
		repo.log.Warn("unknown error while creating new user", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	return nil
}

// Exists checks if user exist with provided id or username. Returns boolean statement that shows existing of user.
func (repo *UserRepository) Exists(ctx context.Context, id string) bool {
	var ok bool
	_ = repo.pool.QueryRow(ctx, `SELECT EXISTS(SELECT * FROM users WHERE id = $1);`, id).Scan(&ok)
	return ok
}

// GetByEmail return user with provided username.
// If user does not exist returns store.ErrNotFound error
func (repo *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (
	u *model.User,
	err error,
) {
	u = new(model.User)

	if err = repo.pool.QueryRow(
		ctx,
		`SELECT x.id, x.email, x.pass FROM users x WHERE x.email = $1;`,
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.Pass,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		repo.log.Debug("unknown error while getting user", TraceError(err)...)
		return nil, store.ErrUnknown
	}

	return u, nil
}
