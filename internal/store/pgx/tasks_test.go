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

func TestTaskRepository_Create_NilTask(t *testing.T) {
	s, td := testStore(t, nil)
	defer td()
	err := s.task.Create(context.Background(), nil)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrNilReference)
	}
}

func TestTaskRepository_Create(t *testing.T) {
	s, td := testStore(t, nil)
	defer td()

	task := &model.Task{
		ID:          uuid.New(),
		Name:        uuid.NewString(),
		Description: uuid.NewString(),
		CreatedAt:   time.Now(),
		CreatedBy:   TestUser1.ID,
		Status:      "NEW",
	}

	ctx := context.Background()

	err := s.task.Create(ctx, task)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrFKViolation)
	}
	require.NoError(t, s.user.Create(ctx, TestUser1))

	assert.False(t, s.task.Exists(ctx, task.ID))
	err = s.task.Create(ctx, task)
	require.NoError(t, err)
	assert.True(t, s.task.Exists(ctx, task.ID))

	err = s.task.Create(ctx, task)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, store.ErrUniqueViolation)
	}
}

func TestTaskRepository_AddToGroup(t *testing.T) {
	s, td := testStore(t, nil)
	defer td()

	task := &model.Task{
		ID:          uuid.New(),
		Name:        uuid.NewString(),
		Description: uuid.NewString(),
		CreatedAt:   time.Now(),
		CreatedBy:   TestUser1.ID,
		Status:      "NEW",
	}

	ctx := context.Background()

	require.NoError(t, s.user.Create(ctx, TestUser1))
	require.NoError(t, s.group.Create(ctx, TestGroup1))
	assert.False(t, s.group.TaskExists(ctx, TestGroup1.ID, task.ID))
	assert.False(t, s.task.Exists(ctx, task.ID))
	require.NoError(t, s.task.Create(ctx, task))

	assert.NoError(t, s.task.AddToGroup(ctx, task.ID, TestGroup1.ID))
	assert.True(t, s.group.TaskExists(ctx, TestGroup1.ID, task.ID))
}
