package postgres

import (
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
