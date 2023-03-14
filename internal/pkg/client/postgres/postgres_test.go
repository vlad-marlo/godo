package postgres

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/config"
	"go.uber.org/fx/fxtest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient_L_NotNil(t *testing.T) {
	cli := &Client{}
	assert.Nil(t, cli.L())
	cli.logger = zap.L()
	assert.Equal(t, cli.logger, cli.L())
}

//goland:noinspection GoNilness
func TestClient_L_Nil(t *testing.T) {
	var cli *Client
	assert.Equal(t, zap.NewNop(), cli.L())
}

func TestClient_P_NotNil(t *testing.T) {
	cli := &Client{}
	assert.Nil(t, cli.P())
	cli.pool = &pgxpool.Pool{}
	assert.Equal(t, cli.pool, cli.P())
}

//goland:noinspection ALL
func TestClient_P_Nil(t *testing.T) {
	var cli *Client
	assert.Nil(t, cli.P())
	assert.NotPanics(t, cli.Close)
}

func TestClient_Close(t *testing.T) {
	tt := []struct {
		name string
		cli  func(t *testing.T) *Client
		want bool
	}{
		{
			name: "positive case #1",
			cli: func(t *testing.T) *Client {
				return &Client{}
			},
			want: false,
		},
		{
			name: "nil client",
			cli: func(t *testing.T) *Client {
				return nil
			},
			want: false,
		},
		{
			name: "nil client",
			cli: func(t *testing.T) *Client {
				return TestClient(t)
			},
			want: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want {
				assert.Panics(t, tc.cli(t).Close)
			} else {
				assert.NotPanics(t, tc.cli(t).Close)
			}
		})
	}
}

func TestNew(t *testing.T) {
	log := zap.L()

	lc := fxtest.NewLifecycle(t)

	cfg := config.New()
	cfg.Postgres = config.Postgres{URI: "postgresql://l:l@l:l/l"}

	cli, err := New(lc, log, cfg)
	require.Error(t, err)
	require.Nil(t, cli)

	cfg.Postgres.URI = fmt.Sprintf("postgresql://%s:%s@%s:5432/%s", uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString())
	cli, err = New(lc, log, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
}
