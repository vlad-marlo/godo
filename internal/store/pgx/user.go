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

// UserRepository ...
type UserRepository struct {
	p *pgxpool.Pool
	l *zap.Logger
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
	if _, err := repo.p.Exec(
		ctx,
		`INSERT INTO users (id, username, pass, is_admin) VALUES ($1, $2, $3, $4);`,
		u.ID,
		u.Name,
		u.Pass,
		u.IsAdmin,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrUserAlreadyExists
			}
		}
		repo.l.Warn("unknown error while creating new user", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}
	return nil
}

// Exists checks if user exist with provided id or username. Returns boolean statement that shows existing of user.
func (repo *UserRepository) Exists(ctx context.Context, user string) bool {
	var ok bool
	_ = repo.p.QueryRow(ctx, `SELECT EXISTS(SELECT * FROM users WHERE id = $1);`, user).Scan(&ok)
	return ok
}

// GetByName return user with provided username.
// If user does not exist returns store.ErrNotFound error
func (repo *UserRepository) GetByName(
	ctx context.Context,
	username string,
) (
	*model.User,
	error,
) {
	u := new(model.User)

	if err := repo.p.QueryRow(
		ctx,
		`SELECT x.id, x.username, x.pass, x.is_admin FROM users x WHERE x.username = $1;`,
		username,
	).Scan(
		&u.ID,
		&u.Name,
		&u.Pass,
		&u.IsAdmin,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		repo.l.Debug("unknown error while getting user", TraceError(err)...)
		return nil, store.ErrUnknown
	}
	return u, nil
}
