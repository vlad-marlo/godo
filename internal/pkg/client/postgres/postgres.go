// This package gives access to db from all internal packages. This package uses singleton object that will be returned
// to user when he wants access to storage.

package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
	"testing"

	pgxzap "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/require"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/config"
)

// Client object gives access to db connection.
type Client struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// New return a singleton client object.
func New(lc fx.Lifecycle, log *zap.Logger, cfg *config.Config) (*Client, error) {
	var pool *pgxpool.Pool

	c, err := pgxpool.ParseConfig(
		cfg.Postgres.URI,
	)
	if err != nil {
		return nil, fmt.Errorf("error while parsing db uri: %w", err)
	}

	var lvl tracelog.LogLevel
	if cfg.Server.IsProd {
		lvl = tracelog.LogLevelError
	} else {
		lvl = tracelog.LogLevelTrace
	}
	c.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgxzap.NewLogger(log),
		LogLevel: lvl,
	}

	// register google uuid support
	c.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, fmt.Errorf("postgres: init pgxpool: %w", err)
	}

	cli := &Client{
		pool:   pool,
		logger: log,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	log.Info("created postgres client")
	return cli, nil
}

// P return pool object with opened connection.
func (c *Client) P() *pgxpool.Pool {
	if c == nil {
		return nil
	}
	return c.pool
}

// Close closes db connection.
func (c *Client) Close() {
	if c == nil {
		return
	}
	if c.logger == nil {
		c.logger = zap.L()
	}
	c.logger.Info("closing poll connection")
	if c.pool != nil {
		c.pool.Close()
	}
}

// L returns prepared logger object.
func (c *Client) L() *zap.Logger {
	if c == nil {
		return zap.L()
	}
	return c.logger
}

// TestClient ...
func TestClient(t testing.TB) *Client {
	t.Helper()
	//TODO: захардкожена переменная окружения мб потом поменять
	dbUri := os.Getenv("TEST_DB_URI")
	c, err := pgxpool.ParseConfig(dbUri)
	require.NoError(t, err)

	c.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(context.Background(), c)
	require.NoError(t, err)
	if err = pool.Ping(context.Background()); err != nil {
		t.Skipf("database is not accessible: %s", err.Error())
	}

	return &Client{
		pool:   pool,
		logger: zap.L(),
	}
}
