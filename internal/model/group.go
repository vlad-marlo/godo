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
		// Name must be unique string. Name will be used
		Name string `json:"name"`
		// Description is additional info about group.
		// For example a Company name or any other meta info.
		Description string `json:"description"`
	}

	// CreateGroupResponse represents group object that was stored in service.
	CreateGroupResponse struct {
		// ID is primary key of group.
		ID uuid.UUID `json:"id"`
		// Name is unique name of group.
		Name string `json:"name"`
		// Description is short info about group.
		Description string `json:"description"`
		// CreatedAt is creation time in UNIX format
		CreatedAt int64 `json:"created-at"`
	}
)
