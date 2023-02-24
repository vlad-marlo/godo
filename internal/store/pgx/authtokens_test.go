package pgx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/pkg/client/postgres"
	"github.com/vlad-marlo/godo/internal/store"
	"testing"
)

func TestTokenRepository_Create(t *testing.T) {
	srv, td := testStore(t, postgres.TestClient(t))
	defer td()

	ctx := context.Background()
	require.NoError(t, srv.User().Create(ctx, TestUser1))
	require.NoError(t, srv.Token().Create(ctx, TestToken1))
	// token already exists
	require.Error(t, srv.Token().Create(ctx, TestToken1))
}

func TestTokenRepository_GetUser(t *testing.T) {
	srv, td := testStore(t, postgres.TestClient(t))
	defer td()
	ctx := context.Background()

	token, err := srv.Token().Get(ctx, TestToken1.Token)
	require.ErrorIs(t, err, store.ErrNotFound)
	assert.Nil(t, token)
	require.NoError(t, srv.User().Create(ctx, TestUser1))
	require.NoError(t, srv.Token().Create(ctx, TestToken1))
	token, err = srv.Token().Get(ctx, TestToken1.Token)
	assert.NotNil(t, token)
	assert.NoError(t, err)
	assert.Equal(t, TestToken1.UserID, token.UserID)
	assert.Equal(t, TestToken1.Expires, token.Expires)

	assert.InDelta(t, TestToken1.ExpiresAt.Second(), token.ExpiresAt.Second(), 1)
}
