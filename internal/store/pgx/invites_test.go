package pgx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/pkg/client/postgres"
	"github.com/vlad-marlo/godo/internal/store"
	"testing"
)

func TestInviteRepository_Create(t *testing.T) {
	ctx := context.Background()
	st, td := testStore(t, nil)
	defer td()
	err := st.invite.Create(context.Background(), TestInvite1, TestRole1, TestGroup1.ID, 1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrFKViolation)

	require.NoError(t, st.user.Create(ctx, TestUser1))
	require.NoError(t, st.group.Create(ctx, TestGroup1))

	require.NoError(t, st.invite.Create(context.Background(), TestInvite1, TestRole1, TestGroup1.ID, 1))
	err = st.invite.Create(ctx, TestInvite1, TestRole1, TestGroup1.ID, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUniqueViolation)
	err = st.invite.Create(ctx, TestInvite2, TestRole1, TestGroup1.ID, -1)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrBadData)
}

func TestInviteRepository_Exists(t *testing.T) {
	st, td := testStore(t, postgres.TestClient(t))
	defer td()
	ctx := context.Background()

	require.False(t, st.invite.Exists(ctx, TestInvite1, TestGroup1.ID))
	require.NoError(t, st.user.Create(ctx, TestUser1))
	require.NoError(t, st.group.Create(ctx, TestGroup1))
	require.NoError(t, st.invite.Create(ctx, TestInvite1, TestRole1, TestGroup1.ID, 1))
	require.True(t, st.invite.Exists(ctx, TestInvite1, TestGroup1.ID))
	require.False(t, st.invite.Exists(ctx, TestInvite2, TestGroup1.ID))
	require.False(t, st.invite.Exists(ctx, TestInvite1, TestGroup2.ID))
}
