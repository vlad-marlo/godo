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
	assert.ErrorIs(t, err, store.ErrBadData)
}

func TestGroupRepository_Create_Positive(t *testing.T) {
	repo, td := testGroup(t)
	defer td()
	user, tdUser := testUsers(t)
	defer tdUser()

	ctx := context.Background()

	require.NoError(t, user.Create(ctx, TestUser1))
	err := repo.Create(ctx, TestGroup1)
	assert.NoError(t, err)
	err = repo.Create(ctx, TestGroup2)
	assert.NoError(t, err)
}

func TestGroupRepository_Create_Negative_Unknown(t *testing.T) {
	cli := BadCli(t)
	grp := NewGroupRepository(cli)
	err := grp.Create(context.Background(), TestGroup1)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUnknown)
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
	assert.ErrorIs(t, err, store.ErrGroupAlreadyExists)
}

func TestGroupRepository_Exists(t *testing.T) {
	grp, usr, td := testGroupUser(t)
	defer td()

	ctx := context.Background()

	err := usr.Create(ctx, TestUser1)
	require.NoError(t, err)

	ok := grp.Exists(ctx, TestGroup1.ID.String())
	require.False(t, ok)
	require.NoError(t, grp.Create(ctx, TestGroup1))
	ok = grp.Exists(ctx, TestGroup1.ID.String())
	assert.True(t, ok)
}

func TestGroupRepository_Get(t *testing.T) {
	grp, usr, td := testGroupUser(t)
	defer td()
	ctx := context.Background()
	err := usr.Create(ctx, TestUser1)
	require.NoError(t, err)

	var group *model.Group
	group, err = grp.Get(ctx, TestGroup1.ID.String())
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrNotFound)
	assert.Nil(t, group)

	err = grp.Create(ctx, TestGroup1)
	assert.NoError(t, err)

	group, err = grp.Get(ctx, TestGroup1.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, group, TestGroup1)

	group, err = grp.Get(ctx, TestGroup2.ID.String())
	assert.Nil(t, group)
	assert.ErrorIs(t, err, store.ErrNotFound)
	assert.Error(t, err)
}

func TestGroupRepository_Get_Negative_BadClient(t *testing.T) {
	cli := BadCli(t)
	grp := NewGroupRepository(cli)
	group, err := grp.Get(context.Background(), TestGroup1.ID.String())
	assert.Nil(t, group)
	assert.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUnknown)
}
