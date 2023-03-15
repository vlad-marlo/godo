package pgx

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"testing"
	"time"
)

func TestGroupRepository_Create_Negative_BadData(t *testing.T) {
	repo, td := testGroup(t)
	defer td()

	ctx := context.Background()
	grp := &model.Group{
		ID:          uuid.New(),
		Name:        "test group",
		Owner:       TestUser1.ID,
		Description: "test description",
		CreatedAt:   time.Now(),
	}
	err := repo.Create(ctx, grp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrFKViolation)
}

func TestGroupRepository_Create_Positive(t *testing.T) {
	s, td := testStore(t, nil)
	defer td()

	ctx := context.Background()

	require.NoError(t, s.user.Create(ctx, TestUser1))
	err := s.group.Create(ctx, TestGroup1)
	assert.NoError(t, err)
	err = s.group.Create(ctx, TestGroup2)
	assert.NoError(t, err)
}

func TestGroupRepository_Create_NilReference(t *testing.T) {
	s, td := testStore(t, nil)
	defer td()

	ctx := context.Background()

	err := s.group.Create(ctx, nil)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrNilReference)
	}
}

func TestGroupRepository_Create_AlreadyExists(t *testing.T) {
	grp, usr, td := testGroupUser(t)
	defer td()
	ctx := context.Background()

	err := usr.Create(ctx, TestUser1)
	require.NoError(t, err)
	require.True(t, usr.Exists(ctx, TestUser1.ID.String()))
	g := TestGroup1
	err = grp.Create(ctx, g)
	require.NoError(t, err)
	err = grp.Create(ctx, g)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUniqueViolation)
}

func TestGroupRepository_Get(t *testing.T) {
	grp, usr, td := testGroupUser(t)
	defer td()
	ctx := context.Background()
	err := usr.Create(ctx, TestUser1)
	require.NoError(t, err)

	var group *model.Group
	group, err = grp.Get(ctx, TestGroup1.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrNotFound)
	assert.Nil(t, group)

	err = grp.Create(ctx, TestGroup1)
	assert.NoError(t, err)

	group, err = grp.Get(ctx, TestGroup1.ID)
	assert.NoError(t, err)
	assert.Equal(t, group, TestGroup1)

	group, err = grp.Get(ctx, TestGroup2.ID)
	assert.Nil(t, group)
	assert.ErrorIs(t, err, store.ErrNotFound)
	assert.Error(t, err)
}
