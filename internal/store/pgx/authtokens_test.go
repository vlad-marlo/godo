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
	err := srv.token.Create(ctx, TestToken1)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrTokenAlreadyExists)
	}
	err = srv.token.Create(ctx, nil)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrNilReference)
	}
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

	require.ErrorIs(t, srv.Token().Create(ctx, TestToken2), store.ErrTokenAlreadyExists)

	assert.NoError(t, srv.Token().Create(ctx, TestToken3))
	token, err = srv.Token().Get(ctx, TestToken3.Token)
	assert.NoError(t, err)
	assert.True(t, TestToken3.ExpiresAt.After(token.ExpiresAt))
}
