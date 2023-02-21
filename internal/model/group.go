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
		CreatedBy   uuid.UUID
		Description string
		CreatedAt   time.Time
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
