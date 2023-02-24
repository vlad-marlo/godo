package model

import (
	"github.com/google/uuid"
	"time"
)

type (
	// Group ...
	Group struct {
		ID          uuid.UUID
		Name        string
		Description string
		CreatedAt   time.Time
		Owner       uuid.UUID
	}

	// CreateGroupRequest ...
	CreateGroupRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// CreateGroupResponse ...
	CreateGroupResponse struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		// time in RFC3339 format
		CreatedAt string `json:"created_at"`
	}
)
