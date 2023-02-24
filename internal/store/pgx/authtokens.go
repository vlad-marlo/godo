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
	"time"
)

type TokenRepository struct {
	p *pgxpool.Pool
	l *zap.Logger
}

// NewTokenRepository is constructor of token repository.
func NewTokenRepository(cli Client) *TokenRepository {
	return &TokenRepository{
		p: cli.P(),
		l: cli.L(),
	}
}

func (repo *TokenRepository) Create(ctx context.Context, token *model.Token) error {
	if _, err := repo.p.Exec(
		ctx,
		`INSERT INTO auth_tokens(user_id, token, expires_at, expires) VALUES ($1, $2, $3, $4);`,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.Expires,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrTokenAlreadyExists
			}
		}
		repo.l.Error("unknown error while creating new user token", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}
	return nil
}

// Get return token with provided body.
func (repo *TokenRepository) Get(ctx context.Context, token string) (*model.Token, error) {
	var t model.Token
	if err := repo.p.QueryRow(
		ctx,
		`SELECT user_id, expires, expires_at FROM auth_tokens WHERE token = $1;`,
		token,
	).Scan(&t.UserID, &t.Expires, &t.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		repo.l.Error("unexpected error while getting user by token", TraceError(err)...)
		return nil, fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	// checking that token is valid - he does not expire or his expiration time was not.
	if time.Now().UTC().After(t.ExpiresAt.UTC()) && t.Expires {
		return nil, store.ErrTokenIsExpired
	}

	return &t, nil
}
