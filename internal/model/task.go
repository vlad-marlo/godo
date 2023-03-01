package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type (
	// Task ...
	Task struct {
		ID          uuid.UUID
		Name        string
		Description string
		CreatedAt   time.Time
		CreatedBy   uuid.UUID
	}
	// TaskCreateRequest ...
	TaskCreateRequest struct {
		// Name is name of task.
		Name string `json:"name"`
		// Description - is verbose info about task. Could be any string.
		Description string `json:"description"`
		// Users - field which relating users to task.
		// If not defined, will create task only for user, who creates this task or for group.
		Users []string `json:"users"`
		// Group - optional filed that show group to which task will be related.
		Group string `json:"group"`
	}
	TaskCreateResponse struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"-"`
		Created     int64     `json:"created-at"`
	}
	// TaskInGroupResponse show short info about task - his name and id.
	// User can check verbose info about task by id.
	TaskInGroupResponse struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
)

// MarshalJSON implements json.Marshaler.
// Used to pass correct time layout to user.
func (task *TaskCreateResponse) MarshalJSON() ([]byte, error) {
	task.Created = task.CreatedAt.Unix()
	return json.Marshal(task)
}

// UnmarshalJSON implements json.Unmarshaler.
// That gives developer flexibility to not thing about correct time layout passed into.
func (task *TaskCreateResponse) UnmarshalJSON(data []byte) (err error) {
	if err = json.Unmarshal(data, &task); err != nil {
		return err
	}
	task.CreatedAt = time.Unix(task.Created, 0)
	return nil
}