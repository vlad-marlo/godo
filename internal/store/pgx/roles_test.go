package pgx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/store"
	"testing"
)

func TestRoleRepository_Create(t *testing.T) {
	str, td := testStore(t, nil)
	defer td()

	ctx := context.Background()
	err := str.role.Create(ctx, nil)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrNilReference)
	}

	err = str.role.Create(ctx, TestRole1)
	require.NoError(t, err)
	err = str.role.Create(ctx, TestRole1)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrUniqueViolation)
	}
}

func TestRoleRepository_Get(t *testing.T) {
	st, td := testStore(t, nil)
	defer td()

	ctx := context.Background()
	err := st.role.Get(ctx, nil)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrNilReference)
	}
	err = st.role.Get(ctx, TestRole1)
	assert.NoError(t, err)
	err = st.role.Get(ctx, TestRole1)
	assert.NoError(t, err)
}
