package model

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTask_MarshalJSON(t *testing.T) {
	b, err := (*Task)(nil).MarshalJSON()
	require.NoError(t, err)
	assert.Nil(t, b)
	err = (*Task)(nil).UnmarshalJSON(nil)
	require.NoError(t, err)

	task := &Task{
		ID:          uuid.New(),
		Name:        "name",
		Description: "description",
		CreatedAt:   time.Now(),
		CreatedBy:   uuid.New(),
		Created:     0,
		Status:      "NEW",
	}
	task.CreatedAt = task.CreatedAt.Round(time.Second)
	b, err = task.MarshalJSON()
	require.NoError(t, err)

	newTask := new(Task)
	err = newTask.UnmarshalJSON(b)
	require.NoError(t, err)

	newTask.CreatedAt = newTask.CreatedAt.Round(time.Second)
	require.Equal(t, task, newTask)
}
