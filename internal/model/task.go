package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type (
	// Task ...
	Task struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"-"`
		CreatedBy   uuid.UUID `json:"created-by"`
		Status      string    `json:"status"`
	}
	// TaskCreateRequest ...
	TaskCreateRequest struct {
		// Name is name of task.
		Name string `json:"name"`
		// Description - is verbose info about task. Could be any string.
		Description string `json:"description"`
		// Users - field which relating users to task.
		// If not defined, will create task only for user, who creates this task or for group.
		Users []uuid.UUID `json:"users"`
		// Group - optional filed that show group to which task will be related.
		Group *uuid.UUID `json:"group"`
	}
	// GetTasksResponse ...
	GetTasksResponse struct {
		Count int     `json:"count"`
		Tasks []*Task `json:"tasks"`
	}
)

// MarshalJSON implements json.Marshaler.
// Used to pass correct time layout to user.
func (task *Task) MarshalJSON() ([]byte, error) {
	if task == nil {
		return nil, nil
	}
	type TransactionAlias Task

	aliasValue := &struct {
		*TransactionAlias
		Created int64 `json:"created-at"`
	}{
		// задаём указатель на целевой объект
		TransactionAlias: (*TransactionAlias)(task),
		Created:          task.CreatedAt.Unix(),
		// вызываем стандартный Unmarshal
	}

	return json.Marshal(aliasValue)
}

// UnmarshalJSON implements json.Unmarshaler.
// That gives developer flexibility to not thing about correct time layout passed into.
func (task *Task) UnmarshalJSON(data []byte) (err error) {
	if task == nil {
		return nil
	}
	type TaskAlias Task
	alias := &struct {
		*TaskAlias
		Created int64 `json:"created-at"`
	}{
		TaskAlias: (*TaskAlias)(task),
	}
	if err = json.Unmarshal(data, alias); err != nil {
		return err
	}

	task.CreatedAt = time.Unix(alias.Created, 0)
	return nil
}
